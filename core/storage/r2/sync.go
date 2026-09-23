package r2

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/dhowden/tag"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/model/id"
	"github.com/navidrome/navidrome/persistence/postgres"
	"github.com/navidrome/navidrome/persistence/postgres/pgdb"
)

var audioExtensions = map[string]bool{
	".mp3":  true,
	".flac": true,
	".m4a":  true,
	".aac":  true,
	".ogg":  true,
	".opus": true,
	".wav":  true,
	".aiff": true,
	".aif":  true,
	".wma":  true,
}

// SyncResult captures the summary of an R2 library synchronization.
type SyncResult struct {
	TotalObjects int           `json:"totalObjects"`
	AudioFiles   int           `json:"audioFiles"`
	Processed    int           `json:"processed"`
	Errors       int           `json:"errors"`
	Duration     time.Duration `json:"duration"`
}

// SyncR2 scans a Cloudflare R2 bucket for audio files, extracts tags via range requests,
// and upserts metadata into Neon/PostgreSQL.
func SyncR2(ctx context.Context, store *postgres.PostgresStore, cfg Config) (*SyncResult, error) {
	start := time.Now()
	res := &SyncResult{}

	client, err := NewS3Client(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating r2 client: %w", err)
	}

	queries := store.Queries()

	// Ensure library exists
	libPath := fmt.Sprintf("r2://%s", cfg.Bucket)
	lib, err := queries.GetLibraryByID(ctx, 1)
	if err != nil {
		lib, err = queries.UpsertLibrary(ctx, pgdb.UpsertLibraryParams{
			ID:              1,
			Name:            "Cloudflare R2 Music",
			Path:            libPath,
			RemotePath:      "",
			DefaultNewUsers: true,
			CreatedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("ensuring default library: %w", err)
		}
	}

	_ = queries.ScanBegin(ctx, pgdb.ScanBeginParams{
		ID:                 lib.ID,
		FullScanInProgress: true,
	})
	defer func() {
		_ = queries.ScanEnd(ctx, lib.ID)
		_ = queries.RefreshLibraryStats(ctx, lib.ID)
	}()

	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(cfg.Bucket),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return res, fmt.Errorf("listing r2 objects: %w", err)
		}

		res.TotalObjects += len(page.Contents)

		for _, obj := range page.Contents {
			key := *obj.Key
			ext := strings.ToLower(filepath.Ext(key))
			if !audioExtensions[ext] {
				continue
			}

			res.AudioFiles++

			if err := processR2AudioObject(ctx, client, queries, cfg.Bucket, key, obj, lib.ID); err != nil {
				log.Error(ctx, "Error processing R2 audio object", "key", key, err)
				res.Errors++
			} else {
				res.Processed++
			}
		}
	}

	res.Duration = time.Since(start)
	return res, nil
}

func processR2AudioObject(
	ctx context.Context,
	client *s3.Client,
	queries *pgdb.Queries,
	bucket, key string,
	obj types.Object,
	libraryID int32,
) error {
	f, err := newR2File(ctx, client, bucket, key)
	if err != nil {
		return fmt.Errorf("opening r2 file for tag reading: %w", err)
	}
	defer f.Close()

	// 1. Attempt to read ID3 / Vorbis tags from header bytes
	parsedTag, _ := tag.ReadFrom(f)

	// 2. Extract metadata with filename/directory fallback
	title, artist, album, albumArtist, year, trackNum, discNum, genre := extractMetadata(key, parsedTag)

	// 3. Generate deterministic Base62 IDs
	artistID := id.NewHash("artist", strings.ToLower(artist))
	albumArtistID := id.NewHash("artist", strings.ToLower(albumArtist))
	albumID := id.NewHash("album", albumArtistID, strings.ToLower(album))
	mediaFileID := id.NewHash("track", bucket, key)

	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	modTime := now
	if obj.LastModified != nil {
		modTime = pgtype.Timestamptz{Time: *obj.LastModified, Valid: true}
	}

	// 4. Upsert Artist
	_, err = queries.UpsertArtist(ctx, pgdb.UpsertArtistParams{
		ID:              artistID,
		Name:            artist,
		SortArtistName:  artist,
		OrderArtistName: artist,
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	if err != nil {
		return fmt.Errorf("upserting artist %q: %w", artist, err)
	}

	_ = queries.AddLibraryArtist(ctx, pgdb.AddLibraryArtistParams{
		LibraryID: libraryID,
		ArtistID:  artistID,
		Stats:     "{}",
	})

	// 5. Upsert Album
	_, err = queries.UpsertAlbum(ctx, pgdb.UpsertAlbumParams{
		ID:                   albumID,
		LibraryID:            libraryID,
		Name:                 album,
		ArtistID:             artistID,
		Artist:               artist,
		AlbumArtist:          albumArtist,
		AlbumArtistID:        albumArtistID,
		MinYear:              int32(year),
		MaxYear:              int32(year),
		Genre:                genre,
		SortAlbumName:        album,
		SortAlbumArtistName:  albumArtist,
		OrderAlbumName:       album,
		OrderAlbumArtistName: albumArtist,
		CreatedAt:            now,
		UpdatedAt:            now,
	})
	if err != nil {
		return fmt.Errorf("upserting album %q: %w", album, err)
	}

	_ = queries.AddAlbumArtist(ctx, pgdb.AddAlbumArtistParams{
		AlbumID:  albumID,
		ArtistID: albumArtistID,
		Role:     "albumartist",
		SubRole:  "",
	})

	// 6. Upsert MediaFile
	suffix := strings.TrimPrefix(path.Ext(key), ".")
	size := int64(0)
	if obj.Size != nil {
		size = *obj.Size
	}

	r2Path := fmt.Sprintf("r2://%s/%s", bucket, key)
	_, err = queries.UpsertMediaFile(ctx, pgdb.UpsertMediaFileParams{
		ID:                   mediaFileID,
		LibraryID:            libraryID,
		Path:                 r2Path,
		Title:                title,
		Album:                album,
		Artist:               artist,
		ArtistID:             artistID,
		AlbumArtist:          albumArtist,
		AlbumArtistID:        albumArtistID,
		AlbumID:              albumID,
		TrackNumber:          int32(trackNum),
		DiscNumber:           int32(discNum),
		Year:                 int32(year),
		Size:                 size,
		Suffix:               suffix,
		Genre:                genre,
		SortTitle:            title,
		SortAlbumName:        album,
		SortArtistName:       artist,
		SortAlbumArtistName:  albumArtist,
		OrderTitle:           title,
		OrderAlbumName:       album,
		OrderArtistName:      artist,
		OrderAlbumArtistName: albumArtist,
		Missing:              false,
		BirthTime:            modTime,
		CreatedAt:            now,
		UpdatedAt:            modTime,
	})
	if err != nil {
		return fmt.Errorf("upserting media file %q: %w", title, err)
	}

	_ = queries.AddMediaFileArtist(ctx, pgdb.AddMediaFileArtistParams{
		MediaFileID: mediaFileID,
		ArtistID:    artistID,
		Role:        "artist",
		SubRole:     "",
	})

	return nil
}

func extractMetadata(key string, m tag.Metadata) (title, artist, album, albumArtist string, year, trackNum, discNum int, genre string) {
	// Defaults from file path: e.g., "Artist/Album/01 - Track.mp3"
	clean := strings.Trim(key, "/")
	parts := strings.Split(clean, "/")
	filename := path.Base(clean)
	ext := path.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	title = base
	album = "Unknown Album"
	artist = "Unknown Artist"

	if len(parts) >= 3 {
		artist = parts[len(parts)-3]
		album = parts[len(parts)-2]
	} else if len(parts) == 2 {
		album = parts[0]
	}

	albumArtist = artist

	if m != nil {
		if m.Title() != "" {
			title = m.Title()
		}
		if m.Artist() != "" {
			artist = m.Artist()
		}
		if m.Album() != "" {
			album = m.Album()
		}
		if m.AlbumArtist() != "" {
			albumArtist = m.AlbumArtist()
		} else {
			albumArtist = artist
		}
		if m.Year() > 0 {
			year = m.Year()
		}
		if t, _ := m.Track(); t > 0 {
			trackNum = t
		}
		if d, _ := m.Disc(); d > 0 {
			discNum = d
		}
		if m.Genre() != "" {
			genre = m.Genre()
		}
	} else {
		// Try parsing track number from filename prefix "01 - Title" or "01. Title"
		if num, rest, ok := splitTrackPrefix(base); ok {
			trackNum = num
			title = rest
		}
	}

	return title, artist, album, albumArtist, year, trackNum, discNum, genre
}

func splitTrackPrefix(name string) (int, string, bool) {
	for i, r := range name {
		if r == ' ' || r == '-' || r == '.' || r == '_' {
			if num, err := strconv.Atoi(name[:i]); err == nil {
				rest := strings.Trim(name[i:], " -._")
				return num, rest, true
			}
			break
		}
	}
	return 0, name, false
}

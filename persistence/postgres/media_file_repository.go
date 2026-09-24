package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/deluan/rest"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/model/criteria"
	"github.com/navidrome/navidrome/model/id"
	"github.com/navidrome/navidrome/model/request"
	"github.com/navidrome/navidrome/persistence/postgres/pgdb"
)

type mediaFileRepository struct {
	ctx     context.Context
	queries *pgdb.Queries
	pool    *pgxpool.Pool
}

func NewMediaFileRepository(ctx context.Context, q *pgdb.Queries, p *pgxpool.Pool) model.MediaFileRepository {
	return &mediaFileRepository{
		ctx:     ctx,
		queries: q,
		pool:    p,
	}
}

func toModelMediaFile(m pgdb.MediaFile) *model.MediaFile {
	var bitDepth *int
	if m.BitDepth != nil {
		v := int(*m.BitDepth)
		bitDepth = &v
	}
	var bpm *int
	if m.Bpm != nil {
		v := int(*m.Bpm)
		bpm = &v
	}

	res := &model.MediaFile{
		ID:                   m.ID,
		PID:                  m.Pid,
		LibraryID:            int(m.LibraryID),
		FolderID:             m.FolderID,
		Path:                 m.Path,
		Title:                m.Title,
		Album:                m.Album,
		ArtistID:             m.ArtistID,
		Artist:               m.Artist,
		AlbumArtistID:        m.AlbumArtistID,
		AlbumArtist:          m.AlbumArtist,
		AlbumID:              m.AlbumID,
		HasCoverArt:          m.HasCoverArt,
		TrackNumber:          int(m.TrackNumber),
		DiscNumber:           int(m.DiscNumber),
		DiscSubtitle:         m.DiscSubtitle,
		Year:                 int(m.Year),
		Date:                 m.Date,
		OriginalYear:         int(m.OriginalYear),
		OriginalDate:         m.OriginalDate,
		ReleaseYear:          int(m.ReleaseYear),
		ReleaseDate:          m.ReleaseDate,
		Size:                 m.Size,
		Suffix:               m.Suffix,
		Duration:             m.Duration,
		BitRate:              int(m.BitRate),
		SampleRate:           int(m.SampleRate),
		BitDepth:             bitDepth,
		Channels:             int(m.Channels),
		Codec:                m.Codec,
		ProbeData:            m.ProbeData,
		Genre:                m.Genre,
		SortTitle:            m.SortTitle,
		SortAlbumName:        m.SortAlbumName,
		SortArtistName:       m.SortArtistName,
		SortAlbumArtistName:  m.SortAlbumArtistName,
		OrderTitle:           m.OrderTitle,
		OrderAlbumName:       m.OrderAlbumName,
		OrderArtistName:      m.OrderArtistName,
		OrderAlbumArtistName: m.OrderAlbumArtistName,
		Compilation:          m.Compilation,
		Comment:              m.Comment,
		Lyrics:               m.Lyrics,
		BPM:                  bpm,
		ExplicitStatus:       m.ExplicitStatus,
		CatalogNum:           m.CatalogNum,
		MbzRecordingID:       m.MbzRecordingID,
		MbzReleaseTrackID:    m.MbzReleaseTrackID,
		MbzAlbumID:           m.MbzAlbumID,
		MbzReleaseGroupID:    m.MbzReleaseGroupID,
		MbzArtistID:          m.MbzArtistID,
		MbzAlbumArtistID:     m.MbzAlbumArtistID,
		MbzAlbumType:         m.MbzAlbumType,
		MbzAlbumComment:      m.MbzAlbumComment,
		RGAlbumGain:          m.RgAlbumGain,
		RGAlbumPeak:          m.RgAlbumPeak,
		RGTrackGain:          m.RgTrackGain,
		RGTrackPeak:          m.RgTrackPeak,
		Missing:              m.Missing,
		CreatedAt:            m.CreatedAt.Time,
		UpdatedAt:            m.UpdatedAt.Time,
	}
	if m.BirthTime.Valid {
		res.BirthTime = m.BirthTime.Time
	}
	return res
}

func (r *mediaFileRepository) Get(id string) (*model.MediaFile, error) {
	m, err := r.queries.GetMediaFileByID(r.ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return toModelMediaFile(m), nil
}

func (r *mediaFileRepository) GetWithParticipants(id string) (*model.MediaFile, error) {
	return r.Get(id)
}

const mediaFileSelectColumns = `
	id, pid, library_id, folder_id, path, title, album, artist, artist_id,
	album_artist, album_artist_id, album_id, has_cover_art, track_number, disc_number,
	disc_subtitle, year, date, original_year, original_date, release_year, release_date,
	size, suffix, duration, bit_rate, sample_rate, bit_depth, channels, codec, probe_data,
	genre, sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
	order_title, order_album_name, order_artist_name, order_album_artist_name,
	compilation, comment, lyrics, bpm, explicit_status, catalog_num, mbz_recording_id,
	mbz_release_track_id, mbz_album_id, mbz_release_group_id, mbz_artist_id,
	mbz_album_artist_id, mbz_album_type, mbz_album_comment, rg_album_gain, rg_album_peak,
	rg_track_gain, rg_track_peak, tags, participants, missing, birth_time, created_at, updated_at
`

func (r *mediaFileRepository) GetAll(options ...model.QueryOptions) (model.MediaFiles, error) {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select(mediaFileSelectColumns).
		From("media_file").
		Where(squirrel.Eq{"media_file.missing": false})

	if len(options) > 0 {
		opt := options[0]
		if opt.Filters != nil {
			builder = builder.Where(opt.Filters)
		}
		if opt.Sort != "" {
			sortCol := "sort_title"
			switch strings.ToLower(opt.Sort) {
			case "track_number", "track":
				sortCol = "disc_number ASC, track_number"
			case "title", "name":
				sortCol = "sort_title"
			case "album":
				sortCol = "sort_album_name"
			case "artist":
				sortCol = "sort_artist_name"
			case "year":
				sortCol = "year"
			case "created_at", "recently_added":
				sortCol = "created_at"
			}
			orderDir := "ASC"
			if strings.ToUpper(opt.Order) == "DESC" {
				orderDir = "DESC"
			}
			builder = builder.OrderBy(fmt.Sprintf("%s %s", sortCol, orderDir))
		} else {
			builder = builder.OrderBy("disc_number ASC, track_number ASC, sort_title ASC")
		}
		if opt.Max > 0 {
			builder = builder.Limit(uint64(opt.Max))
		} else {
			builder = builder.Limit(1000)
		}
		if opt.Offset > 0 {
			builder = builder.Offset(uint64(opt.Offset))
		}
	} else {
		builder = builder.OrderBy("disc_number ASC, track_number ASC, sort_title ASC").Limit(1000)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		query = fmt.Sprintf("SELECT %s FROM media_file WHERE missing = FALSE ORDER BY disc_number ASC, track_number ASC LIMIT 1000", mediaFileSelectColumns)
		args = nil
	}

	rows, err := r.pool.Query(r.ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs model.MediaFiles
	for rows.Next() {
		var m pgdb.MediaFile
		if err := rows.Scan(
			&m.ID, &m.Pid, &m.LibraryID, &m.FolderID, &m.Path, &m.Title, &m.Album, &m.Artist, &m.ArtistID,
			&m.AlbumArtist, &m.AlbumArtistID, &m.AlbumID, &m.HasCoverArt, &m.TrackNumber, &m.DiscNumber,
			&m.DiscSubtitle, &m.Year, &m.Date, &m.OriginalYear, &m.OriginalDate, &m.ReleaseYear, &m.ReleaseDate,
			&m.Size, &m.Suffix, &m.Duration, &m.BitRate, &m.SampleRate, &m.BitDepth, &m.Channels, &m.Codec, &m.ProbeData,
			&m.Genre, &m.SortTitle, &m.SortAlbumName, &m.SortArtistName, &m.SortAlbumArtistName,
			&m.OrderTitle, &m.OrderAlbumName, &m.OrderArtistName, &m.OrderAlbumArtistName,
			&m.Compilation, &m.Comment, &m.Lyrics, &m.Bpm, &m.ExplicitStatus, &m.CatalogNum, &m.MbzRecordingID,
			&m.MbzReleaseTrackID, &m.MbzAlbumID, &m.MbzReleaseGroupID, &m.MbzArtistID,
			&m.MbzAlbumArtistID, &m.MbzAlbumType, &m.MbzAlbumComment, &m.RgAlbumGain, &m.RgAlbumPeak,
			&m.RgTrackGain, &m.RgTrackPeak, &m.Tags, &m.Participants, &m.Missing, &m.BirthTime, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		songs = append(songs, *toModelMediaFile(m))
	}
	return songs, nil
}

func (r *mediaFileRepository) CountAll(options ...model.QueryOptions) (int64, error) {
	if len(options) > 0 && options[0].Filters != nil {
		builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
			Select("count(*)").
			From("media_file").
			Where(squirrel.Eq{"media_file.missing": false}).
			Where(options[0].Filters)
		query, args, err := builder.ToSql()
		if err == nil {
			var count int64
			err := r.pool.QueryRow(r.ctx, query, args...).Scan(&count)
			if err == nil {
				return count, nil
			}
		}
	}
	return r.queries.CountMediaFiles(r.ctx)
}

func (r *mediaFileRepository) CountBySuffix(_ ...model.QueryOptions) (map[string]int64, error) {
	rows, err := r.pool.Query(r.ctx, `
		SELECT suffix, count(*)
		FROM media_file
		WHERE missing = FALSE
		GROUP BY suffix
	`)
	if err != nil {
		return map[string]int64{}, nil
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var suffix string
		var count int64
		if err := rows.Scan(&suffix, &count); err == nil {
			counts[suffix] = count
		}
	}
	return counts, nil
}

func (r *mediaFileRepository) Exists(id string) (bool, error) {
	_, err := r.Get(id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *mediaFileRepository) Put(m *model.MediaFile) error {
	var bitDepth *int32
	if m.BitDepth != nil {
		v := int32(*m.BitDepth)
		bitDepth = &v
	}
	var bpm *int32
	if m.BPM != nil {
		v := int32(*m.BPM)
		bpm = &v
	}

	params := pgdb.UpsertMediaFileParams{
		ID:                   m.ID,
		Pid:                  m.PID,
		LibraryID:            int32(m.LibraryID),
		FolderID:             m.FolderID,
		Path:                 m.Path,
		Title:                m.Title,
		Album:                m.Album,
		Artist:               m.Artist,
		ArtistID:             m.ArtistID,
		AlbumArtist:          m.AlbumArtist,
		AlbumArtistID:        m.AlbumArtistID,
		AlbumID:              m.AlbumID,
		HasCoverArt:          m.HasCoverArt,
		TrackNumber:          int32(m.TrackNumber),
		DiscNumber:           int32(m.DiscNumber),
		DiscSubtitle:         m.DiscSubtitle,
		Year:                 int32(m.Year),
		Date:                 m.Date,
		OriginalYear:         int32(m.OriginalYear),
		OriginalDate:         m.OriginalDate,
		ReleaseYear:          int32(m.ReleaseYear),
		ReleaseDate:          m.ReleaseDate,
		Size:                 m.Size,
		Suffix:               m.Suffix,
		Duration:             m.Duration,
		BitRate:              int32(m.BitRate),
		SampleRate:           int32(m.SampleRate),
		BitDepth:             bitDepth,
		Channels:             int32(m.Channels),
		Codec:                m.Codec,
		ProbeData:            m.ProbeData,
		Genre:                m.Genre,
		SortTitle:            m.SortTitle,
		SortAlbumName:        m.SortAlbumName,
		SortArtistName:       m.SortArtistName,
		SortAlbumArtistName:  m.SortAlbumArtistName,
		OrderTitle:           m.OrderTitle,
		OrderAlbumName:       m.OrderAlbumName,
		OrderArtistName:      m.OrderArtistName,
		OrderAlbumArtistName: m.OrderAlbumArtistName,
		Compilation:          m.Compilation,
		Comment:              m.Comment,
		Lyrics:               m.Lyrics,
		Bpm:                  bpm,
		ExplicitStatus:       m.ExplicitStatus,
		CatalogNum:           m.CatalogNum,
		MbzRecordingID:       m.MbzRecordingID,
		MbzReleaseTrackID:    m.MbzReleaseTrackID,
		MbzAlbumID:           m.MbzAlbumID,
		MbzReleaseGroupID:    m.MbzReleaseGroupID,
		MbzArtistID:          m.MbzArtistID,
		MbzAlbumArtistID:     m.MbzAlbumArtistID,
		MbzAlbumType:         m.MbzAlbumType,
		MbzAlbumComment:      m.MbzAlbumComment,
		RgAlbumGain:          m.RGAlbumGain,
		RgAlbumPeak:          m.RGAlbumPeak,
		RgTrackGain:          m.RGTrackGain,
		RgTrackPeak:          m.RGTrackPeak,
		Tags:                 "",
		Participants:         "",
		Missing:              m.Missing,
		BirthTime:            pgtype.Timestamptz{Time: m.BirthTime, Valid: !m.BirthTime.IsZero()},
		CreatedAt:            pgtype.Timestamptz{Time: m.CreatedAt, Valid: !m.CreatedAt.IsZero()},
		UpdatedAt:            pgtype.Timestamptz{Time: m.UpdatedAt, Valid: !m.UpdatedAt.IsZero()},
	}

	_, err := r.queries.UpsertMediaFile(r.ctx, params)
	return err
}

func (r *mediaFileRepository) UpdateProbeData(id string, data string) error {
	_, err := r.pool.Exec(r.ctx, `UPDATE media_file SET probe_data = $1 WHERE id = $2`, data, id)
	return err
}

func (r *mediaFileRepository) GetRandom(options ...model.QueryOptions) (model.MediaFiles, error) {
	limit := 50
	if len(options) > 0 && options[0].Max > 0 {
		limit = options[0].Max
	}
	query := fmt.Sprintf(`SELECT %s FROM media_file WHERE missing = FALSE ORDER BY RANDOM() LIMIT $1`, mediaFileSelectColumns)
	rows, err := r.pool.Query(r.ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs model.MediaFiles
	for rows.Next() {
		var m pgdb.MediaFile
		if err := rows.Scan(
			&m.ID, &m.Pid, &m.LibraryID, &m.FolderID, &m.Path, &m.Title, &m.Album, &m.Artist, &m.ArtistID,
			&m.AlbumArtist, &m.AlbumArtistID, &m.AlbumID, &m.HasCoverArt, &m.TrackNumber, &m.DiscNumber,
			&m.DiscSubtitle, &m.Year, &m.Date, &m.OriginalYear, &m.OriginalDate, &m.ReleaseYear, &m.ReleaseDate,
			&m.Size, &m.Suffix, &m.Duration, &m.BitRate, &m.SampleRate, &m.BitDepth, &m.Channels, &m.Codec, &m.ProbeData,
			&m.Genre, &m.SortTitle, &m.SortAlbumName, &m.SortArtistName, &m.SortAlbumArtistName,
			&m.OrderTitle, &m.OrderAlbumName, &m.OrderArtistName, &m.OrderAlbumArtistName,
			&m.Compilation, &m.Comment, &m.Lyrics, &m.Bpm, &m.ExplicitStatus, &m.CatalogNum, &m.MbzRecordingID,
			&m.MbzReleaseTrackID, &m.MbzAlbumID, &m.MbzReleaseGroupID, &m.MbzArtistID,
			&m.MbzAlbumArtistID, &m.MbzAlbumType, &m.MbzAlbumComment, &m.RgAlbumGain, &m.RgAlbumPeak,
			&m.RgTrackGain, &m.RgTrackPeak, &m.Tags, &m.Participants, &m.Missing, &m.BirthTime, &m.CreatedAt, &m.UpdatedAt,
		); err == nil {
			songs = append(songs, *toModelMediaFile(m))
		}
	}
	return songs, nil
}

func (r *mediaFileRepository) GetAllByTags(_ model.TagName, _ []string, options ...model.QueryOptions) (model.MediaFiles, error) {
	return r.GetAll(options...)
}

func (r *mediaFileRepository) MatchesCriteria(_ string, _ criteria.Criteria) (bool, error) {
	return true, nil
}

func (r *mediaFileRepository) GetCursor(options ...model.QueryOptions) (model.MediaFileCursor, error) {
	songs, err := r.GetAll(options...)
	if err != nil {
		return nil, err
	}
	return func(yield func(model.MediaFile, error) bool) {
		for _, s := range songs {
			if !yield(s, nil) {
				return
			}
		}
	}, nil
}

func (r *mediaFileRepository) GetAlbumIDsByFolder(_ model.Library, _ ...string) ([]string, error) {
	return nil, nil
}

func (r *mediaFileRepository) GetCursorWithArtwork(options ...model.QueryOptions) (model.MediaFileCursor, error) {
	return r.GetCursor(options...)
}

func (r *mediaFileRepository) Delete(id string) error {
	return r.queries.DeleteMediaFile(r.ctx, id)
}

func (r *mediaFileRepository) DeleteMissing(_ []string) error {
	return nil
}

func (r *mediaFileRepository) DeleteAllMissing() (int64, error) {
	return 0, nil
}

func (r *mediaFileRepository) FindByPaths(_ []string) (model.MediaFiles, error) {
	return nil, nil
}

func (r *mediaFileRepository) ReassignReferences(_, _ string) error {
	return nil
}

func (r *mediaFileRepository) MarkMissing(_ bool, _ ...*model.MediaFile) error {
	return nil
}

func (r *mediaFileRepository) MarkMissingByFolder(_ bool, _ ...string) error {
	return nil
}

func (r *mediaFileRepository) GetMissingAndMatching(_ int) (model.MediaFileCursor, error) {
	return func(yield func(model.MediaFile, error) bool) {}, nil
}

func (r *mediaFileRepository) FindRecentFilesByMBZTrackID(_ model.MediaFile, _ time.Time) (model.MediaFiles, error) {
	return nil, nil
}

func (r *mediaFileRepository) FindRecentFilesByProperties(_ model.MediaFile, _ time.Time) (model.MediaFiles, error) {
	return nil, nil
}

func (r *mediaFileRepository) IncPlayCount(itemID string, ts time.Time) error {
	user, ok := request.UserFrom(r.ctx)
	if !ok {
		return nil
	}
	annID := id.NewHash(user.ID + itemID + "media_file")
	_, err := r.pool.Exec(r.ctx, `
		INSERT INTO annotation (ann_id, user_id, item_id, item_type, play_count, play_date)
		VALUES ($1, $2, $3, 'media_file', 1, $4)
		ON CONFLICT (user_id, item_id, item_type)
		DO UPDATE SET play_count = annotation.play_count + 1, play_date = EXCLUDED.play_date
	`, annID, user.ID, itemID, ts)
	return err
}

func (r *mediaFileRepository) SetStar(starred bool, itemIDs ...string) error {
	user, ok := request.UserFrom(r.ctx)
	if !ok {
		return nil
	}
	now := time.Now()
	for _, itemID := range itemIDs {
		annID := id.NewHash(user.ID + itemID + "media_file")
		_, _ = r.pool.Exec(r.ctx, `
			INSERT INTO annotation (ann_id, user_id, item_id, item_type, starred, starred_at)
			VALUES ($1, $2, $3, 'media_file', $4, $5)
			ON CONFLICT (user_id, item_id, item_type)
			DO UPDATE SET starred = EXCLUDED.starred, starred_at = EXCLUDED.starred_at
		`, annID, user.ID, itemID, starred, now)
	}
	return nil
}

func (r *mediaFileRepository) SetRating(rating int, itemID string) error {
	user, ok := request.UserFrom(r.ctx)
	if !ok {
		return nil
	}
	now := time.Now()
	annID := id.NewHash(user.ID + itemID + "media_file")
	_, err := r.pool.Exec(r.ctx, `
		INSERT INTO annotation (ann_id, user_id, item_id, item_type, rating, rated_at)
		VALUES ($1, $2, $3, 'media_file', $4, $5)
		ON CONFLICT (user_id, item_id, item_type)
		DO UPDATE SET rating = EXCLUDED.rating, rated_at = EXCLUDED.rated_at
	`, annID, user.ID, itemID, rating, now)
	return err
}

func (r *mediaFileRepository) ReassignAnnotation(prevID string, newID string) error {
	_, err := r.pool.Exec(r.ctx, `
		UPDATE annotation SET item_id = $1 WHERE item_id = $2 AND item_type = 'media_file'
	`, newID, prevID)
	return err
}

func (r *mediaFileRepository) AddBookmark(_, _ string, _ int64) error {
	return nil
}

func (r *mediaFileRepository) DeleteBookmark(_ string) error {
	return nil
}

func (r *mediaFileRepository) GetBookmarks() (model.Bookmarks, error) {
	return model.Bookmarks{}, nil
}

func (r *mediaFileRepository) Search(q string, options ...model.QueryOptions) (model.MediaFiles, error) {
	limit := 50
	if len(options) > 0 && options[0].Max > 0 {
		limit = options[0].Max
	}
	pattern := "%" + q + "%"
	query := fmt.Sprintf(`
		SELECT %s
		FROM media_file
		WHERE missing = FALSE AND (title ILIKE $1 OR album ILIKE $1 OR artist ILIKE $1)
		ORDER BY sort_title ASC
		LIMIT $2
	`, mediaFileSelectColumns)

	rows, err := r.pool.Query(r.ctx, query, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs model.MediaFiles
	for rows.Next() {
		var m pgdb.MediaFile
		if err := rows.Scan(
			&m.ID, &m.Pid, &m.LibraryID, &m.FolderID, &m.Path, &m.Title, &m.Album, &m.Artist, &m.ArtistID,
			&m.AlbumArtist, &m.AlbumArtistID, &m.AlbumID, &m.HasCoverArt, &m.TrackNumber, &m.DiscNumber,
			&m.DiscSubtitle, &m.Year, &m.Date, &m.OriginalYear, &m.OriginalDate, &m.ReleaseYear, &m.ReleaseDate,
			&m.Size, &m.Suffix, &m.Duration, &m.BitRate, &m.SampleRate, &m.BitDepth, &m.Channels, &m.Codec, &m.ProbeData,
			&m.Genre, &m.SortTitle, &m.SortAlbumName, &m.SortArtistName, &m.SortAlbumArtistName,
			&m.OrderTitle, &m.OrderAlbumName, &m.OrderArtistName, &m.OrderAlbumArtistName,
			&m.Compilation, &m.Comment, &m.Lyrics, &m.Bpm, &m.ExplicitStatus, &m.CatalogNum, &m.MbzRecordingID,
			&m.MbzReleaseTrackID, &m.MbzAlbumID, &m.MbzReleaseGroupID, &m.MbzArtistID,
			&m.MbzAlbumArtistID, &m.MbzAlbumType, &m.MbzAlbumComment, &m.RgAlbumGain, &m.RgAlbumPeak,
			&m.RgTrackGain, &m.RgTrackPeak, &m.Tags, &m.Participants, &m.Missing, &m.BirthTime, &m.CreatedAt, &m.UpdatedAt,
		); err == nil {
			songs = append(songs, *toModelMediaFile(m))
		}
	}
	return songs, nil
}

// REST Repository integration
func (r *mediaFileRepository) Count(_ ...rest.QueryOptions) (int64, error) {
	return r.CountAll()
}

func (r *mediaFileRepository) Read(id string) (any, error) {
	return r.Get(id)
}

func (r *mediaFileRepository) ReadAll(options ...rest.QueryOptions) (any, error) {
	var qo []model.QueryOptions
	if len(options) > 0 {
		qo = append(qo, model.QueryOptions{
			Sort:   options[0].Sort,
			Order:  options[0].Order,
			Max:    options[0].Max,
			Offset: options[0].Offset,
		})
	}
	return r.GetAll(qo...)
}

func (r *mediaFileRepository) EntityName() string {
	return "mediafile"
}

func (r *mediaFileRepository) NewInstance() any {
	return &model.MediaFile{}
}

var _ model.MediaFileRepository = (*mediaFileRepository)(nil)
var _ rest.Repository = (*mediaFileRepository)(nil)

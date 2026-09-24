package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/deluan/rest"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/model/id"
	"github.com/navidrome/navidrome/model/request"
	"github.com/navidrome/navidrome/persistence/postgres/pgdb"
)

type albumRepository struct {
	ctx     context.Context
	queries *pgdb.Queries
	pool    *pgxpool.Pool
}

func NewAlbumRepository(ctx context.Context, q *pgdb.Queries, p *pgxpool.Pool) model.AlbumRepository {
	return &albumRepository{
		ctx:     ctx,
		queries: q,
		pool:    p,
	}
}

func toModelAlbum(a pgdb.Album) *model.Album {
	res := &model.Album{
		ID:                   a.ID,
		LibraryID:            int(a.LibraryID),
		Name:                 a.Name,
		AlbumArtist:          a.AlbumArtist,
		AlbumArtistID:        a.AlbumArtistID,
		EmbedArtPath:         a.EmbedArtPath,
		MaxYear:              int(a.MaxYear),
		MinYear:              int(a.MinYear),
		Date:                 a.Date,
		MaxOriginalYear:      int(a.MaxOriginalYear),
		MinOriginalYear:      int(a.MinOriginalYear),
		OriginalDate:         a.OriginalDate,
		ReleaseDate:          a.ReleaseDate,
		Compilation:          a.Compilation,
		Comment:              a.Comment,
		SongCount:            int(a.SongCount),
		Duration:             a.Duration,
		Size:                 a.Size,
		SortAlbumName:        a.SortAlbumName,
		SortAlbumArtistName:  a.SortAlbumArtistName,
		OrderAlbumName:       a.OrderAlbumName,
		OrderAlbumArtistName: a.OrderAlbumArtistName,
		CatalogNum:           a.CatalogNum,
		MbzAlbumID:           a.MbzAlbumID,
		MbzAlbumArtistID:     a.MbzAlbumArtistID,
		MbzAlbumType:         a.MbzAlbumType,
		MbzAlbumComment:      a.MbzAlbumComment,
		MbzReleaseGroupID:    a.MbzReleaseGroupID,
		ExplicitStatus:       a.ExplicitStatus,
		RGAlbumGain:          a.RgAlbumGain,
		RGAlbumPeak:          a.RgAlbumPeak,
		Description:          a.Description,
		SmallImageUrl:        a.SmallImageUrl,
		MediumImageUrl:       a.MediumImageUrl,
		LargeImageUrl:        a.LargeImageUrl,
		ExternalUrl:          a.ExternalUrl,
		Genre:                a.Genre,
		Missing:              a.Missing,
		CreatedAt:            a.CreatedAt.Time,
		UpdatedAt:            a.UpdatedAt.Time,
	}
	if a.ExternalInfoUpdatedAt.Valid {
		t := a.ExternalInfoUpdatedAt.Time
		res.ExternalInfoUpdatedAt = &t
	}
	if a.ImportedAt.Valid {
		res.ImportedAt = a.ImportedAt.Time
	}
	return res
}

func (r *albumRepository) Get(id string) (*model.Album, error) {
	a, err := r.queries.GetAlbumByID(r.ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return toModelAlbum(a), nil
}

func (r *albumRepository) GetAll(options ...model.QueryOptions) (model.Albums, error) {
	limit := int32(1000)
	offset := int32(0)
	orderCol := "sort_album_name"
	orderDir := "ASC"

	if len(options) > 0 {
		opt := options[0]
		if opt.Max > 0 {
			limit = int32(opt.Max)
		}
		if opt.Offset > 0 {
			offset = int32(opt.Offset)
		}
		if opt.Sort != "" {
			switch strings.ToLower(opt.Sort) {
			case "recently_added", "created_at":
				orderCol = "created_at"
			case "name", "title":
				orderCol = "name"
			case "artist":
				orderCol = "artist"
			case "year", "min_year":
				orderCol = "min_year"
			case "duration":
				orderCol = "duration"
			case "song_count":
				orderCol = "song_count"
			case "starred_at", "play_date":
				orderCol = "updated_at"
			default:
				orderCol = "sort_album_name"
			}
			if strings.ToUpper(opt.Order) == "DESC" {
				orderDir = "DESC"
			}
		}
	}

	query := fmt.Sprintf(`
		SELECT id, library_id, name, artist_id, artist, album_artist, album_artist_id,
		       embed_art_path, cover_art_path, cover_art_id, max_year, min_year, date,
		       max_original_year, min_original_year, original_date, release_date, compilation,
		       comment, song_count, duration, size, discs, sort_album_name, sort_album_artist_name,
		       order_album_name, order_album_artist_name, catalog_num, mbz_album_id, mbz_album_artist_id,
		       mbz_album_type, mbz_album_comment, mbz_release_group_id, folder_ids, explicit_status,
		       rg_album_gain, rg_album_peak, description, small_image_url, medium_image_url,
		       large_image_url, external_url, external_info_updated_at, genre, tags, missing,
		       imported_at, created_at, updated_at
		FROM album
		WHERE missing = FALSE
		ORDER BY %s %s, name ASC
		LIMIT $1 OFFSET $2
	`, orderCol, orderDir)

	rows, err := r.pool.Query(r.ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums model.Albums
	for rows.Next() {
		var a pgdb.Album
		if err := rows.Scan(
			&a.ID, &a.LibraryID, &a.Name, &a.ArtistID, &a.Artist, &a.AlbumArtist, &a.AlbumArtistID,
			&a.EmbedArtPath, &a.CoverArtPath, &a.CoverArtID, &a.MaxYear, &a.MinYear, &a.Date,
			&a.MaxOriginalYear, &a.MinOriginalYear, &a.OriginalDate, &a.ReleaseDate, &a.Compilation,
			&a.Comment, &a.SongCount, &a.Duration, &a.Size, &a.Discs, &a.SortAlbumName, &a.SortAlbumArtistName,
			&a.OrderAlbumName, &a.OrderAlbumArtistName, &a.CatalogNum, &a.MbzAlbumID, &a.MbzAlbumArtistID,
			&a.MbzAlbumType, &a.MbzAlbumComment, &a.MbzReleaseGroupID, &a.FolderIds, &a.ExplicitStatus,
			&a.RgAlbumGain, &a.RgAlbumPeak, &a.Description, &a.SmallImageUrl, &a.MediumImageUrl,
			&a.LargeImageUrl, &a.ExternalUrl, &a.ExternalInfoUpdatedAt, &a.Genre, &a.Tags, &a.Missing,
			&a.ImportedAt, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		albums = append(albums, *toModelAlbum(a))
	}
	return albums, nil
}

func (r *albumRepository) CountAll(_ ...model.QueryOptions) (int64, error) {
	return r.queries.CountAlbums(r.ctx)
}

func (r *albumRepository) Exists(id string) (bool, error) {
	_, err := r.Get(id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *albumRepository) Put(a *model.Album) error {
	params := pgdb.UpsertAlbumParams{
		ID:                   a.ID,
		LibraryID:            int32(a.LibraryID),
		Name:                 a.Name,
		ArtistID:             a.AlbumArtistID,
		Artist:               a.AlbumArtist,
		AlbumArtist:          a.AlbumArtist,
		AlbumArtistID:        a.AlbumArtistID,
		EmbedArtPath:         a.EmbedArtPath,
		CoverArtPath:         a.EmbedArtPath,
		CoverArtID:           a.CoverArtID().String(),
		MaxYear:              int32(a.MaxYear),
		MinYear:              int32(a.MinYear),
		Date:                 a.Date,
		MaxOriginalYear:      int32(a.MaxOriginalYear),
		MinOriginalYear:      int32(a.MinOriginalYear),
		OriginalDate:         a.OriginalDate,
		ReleaseDate:          a.ReleaseDate,
		Compilation:          a.Compilation,
		Comment:              a.Comment,
		SongCount:            int32(a.SongCount),
		Duration:             a.Duration,
		Size:                 a.Size,
		Discs:                "{}",
		SortAlbumName:        a.SortAlbumName,
		SortAlbumArtistName:  a.SortAlbumArtistName,
		OrderAlbumName:       a.OrderAlbumName,
		OrderAlbumArtistName: a.OrderAlbumArtistName,
		CatalogNum:           a.CatalogNum,
		MbzAlbumID:           a.MbzAlbumID,
		MbzAlbumArtistID:     a.MbzAlbumArtistID,
		MbzAlbumType:         a.MbzAlbumType,
		MbzAlbumComment:      a.MbzAlbumComment,
		MbzReleaseGroupID:    a.MbzReleaseGroupID,
		FolderIds:            "",
		ExplicitStatus:       a.ExplicitStatus,
		RgAlbumGain:          a.RGAlbumGain,
		RgAlbumPeak:          a.RGAlbumPeak,
		Description:          a.Description,
		SmallImageUrl:        a.SmallImageUrl,
		MediumImageUrl:       a.MediumImageUrl,
		LargeImageUrl:        a.LargeImageUrl,
		ExternalUrl:          a.ExternalUrl,
		Genre:                a.Genre,
		Tags:                 "",
		Missing:              a.Missing,
		ImportedAt:           pgtype.Timestamptz{Time: a.ImportedAt, Valid: !a.ImportedAt.IsZero()},
		CreatedAt:            pgtype.Timestamptz{Time: a.CreatedAt, Valid: !a.CreatedAt.IsZero()},
		UpdatedAt:            pgtype.Timestamptz{Time: a.UpdatedAt, Valid: !a.UpdatedAt.IsZero()},
	}
	_, err := r.queries.UpsertAlbum(r.ctx, params)
	return err
}

func (r *albumRepository) UpdateExternalInfo(a *model.Album) error {
	return r.Put(a)
}

func (r *albumRepository) GetSoleAlbumArtistIDsInSubtrees(_ model.Library, _ ...string) ([]string, error) {
	return nil, nil
}

func (r *albumRepository) GetCursor(options ...model.QueryOptions) (model.AlbumCursor, error) {
	albums, err := r.GetAll(options...)
	if err != nil {
		return nil, err
	}
	return func(yield func(model.Album, error) bool) {
		for _, a := range albums {
			if !yield(a, nil) {
				return
			}
		}
	}, nil
}

func (r *albumRepository) GetYears(_ ...int) ([]int, error) {
	rows, err := r.pool.Query(r.ctx, `
		SELECT DISTINCT min_year
		FROM album
		WHERE missing = FALSE AND min_year > 0
		ORDER BY min_year DESC
	`)
	if err != nil {
		return []int{}, nil
	}
	defer rows.Close()
	var years []int
	for rows.Next() {
		var y int
		if err := rows.Scan(&y); err == nil {
			years = append(years, y)
		}
	}
	return years, nil
}

func (r *albumRepository) Touch(_ ...string) error {
	return nil
}

func (r *albumRepository) TouchByMissingFolder() (int64, error) {
	return 0, nil
}

func (r *albumRepository) GetTouchedAlbums(_ int) (model.AlbumCursor, error) {
	return func(yield func(model.Album, error) bool) {}, nil
}

func (r *albumRepository) RefreshPlayCounts() (int64, error) {
	return 0, nil
}

func (r *albumRepository) CopyAttributes(_, _ string, _ ...string) error {
	return nil
}

func (r *albumRepository) IncPlayCount(itemID string, ts time.Time) error {
	user, ok := request.UserFrom(r.ctx)
	if !ok {
		return nil
	}
	annID := id.NewHash(user.ID + itemID + "album")
	_, err := r.pool.Exec(r.ctx, `
		INSERT INTO annotation (ann_id, user_id, item_id, item_type, play_count, play_date)
		VALUES ($1, $2, $3, 'album', 1, $4)
		ON CONFLICT (user_id, item_id, item_type)
		DO UPDATE SET play_count = annotation.play_count + 1, play_date = EXCLUDED.play_date
	`, annID, user.ID, itemID, ts)
	return err
}

func (r *albumRepository) SetStar(starred bool, itemIDs ...string) error {
	user, ok := request.UserFrom(r.ctx)
	if !ok {
		return nil
	}
	now := time.Now()
	for _, itemID := range itemIDs {
		annID := id.NewHash(user.ID + itemID + "album")
		_, _ = r.pool.Exec(r.ctx, `
			INSERT INTO annotation (ann_id, user_id, item_id, item_type, starred, starred_at)
			VALUES ($1, $2, $3, 'album', $4, $5)
			ON CONFLICT (user_id, item_id, item_type)
			DO UPDATE SET starred = EXCLUDED.starred, starred_at = EXCLUDED.starred_at
		`, annID, user.ID, itemID, starred, now)
	}
	return nil
}

func (r *albumRepository) SetRating(rating int, itemID string) error {
	user, ok := request.UserFrom(r.ctx)
	if !ok {
		return nil
	}
	now := time.Now()
	annID := id.NewHash(user.ID + itemID + "album")
	_, err := r.pool.Exec(r.ctx, `
		INSERT INTO annotation (ann_id, user_id, item_id, item_type, rating, rated_at)
		VALUES ($1, $2, $3, 'album', $4, $5)
		ON CONFLICT (user_id, item_id, item_type)
		DO UPDATE SET rating = EXCLUDED.rating, rated_at = EXCLUDED.rated_at
	`, annID, user.ID, itemID, rating, now)
	return err
}

func (r *albumRepository) ReassignAnnotation(prevID string, newID string) error {
	_, err := r.pool.Exec(r.ctx, `
		UPDATE annotation SET item_id = $1 WHERE item_id = $2 AND item_type = 'album'
	`, newID, prevID)
	return err
}

func (r *albumRepository) Search(q string, options ...model.QueryOptions) (model.Albums, error) {
	limit := int32(50)
	if len(options) > 0 && options[0].Max > 0 {
		limit = int32(options[0].Max)
	}
	pattern := "%" + q + "%"
	query := `
		SELECT id, library_id, name, artist_id, artist, album_artist, album_artist_id,
		       embed_art_path, cover_art_path, cover_art_id, max_year, min_year, date,
		       max_original_year, min_original_year, original_date, release_date, compilation,
		       comment, song_count, duration, size, discs, sort_album_name, sort_album_artist_name,
		       order_album_name, order_album_artist_name, catalog_num, mbz_album_id, mbz_album_artist_id,
		       mbz_album_type, mbz_album_comment, mbz_release_group_id, folder_ids, explicit_status,
		       rg_album_gain, rg_album_peak, description, small_image_url, medium_image_url,
		       large_image_url, external_url, external_info_updated_at, genre, tags, missing,
		       imported_at, created_at, updated_at
		FROM album
		WHERE missing = FALSE AND (name ILIKE $1 OR artist ILIKE $1)
		ORDER BY sort_album_name ASC
		LIMIT $2
	`
	rows, err := r.pool.Query(r.ctx, query, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums model.Albums
	for rows.Next() {
		var a pgdb.Album
		if err := rows.Scan(
			&a.ID, &a.LibraryID, &a.Name, &a.ArtistID, &a.Artist, &a.AlbumArtist, &a.AlbumArtistID,
			&a.EmbedArtPath, &a.CoverArtPath, &a.CoverArtID, &a.MaxYear, &a.MinYear, &a.Date,
			&a.MaxOriginalYear, &a.MinOriginalYear, &a.OriginalDate, &a.ReleaseDate, &a.Compilation,
			&a.Comment, &a.SongCount, &a.Duration, &a.Size, &a.Discs, &a.SortAlbumName, &a.SortAlbumArtistName,
			&a.OrderAlbumName, &a.OrderAlbumArtistName, &a.CatalogNum, &a.MbzAlbumID, &a.MbzAlbumArtistID,
			&a.MbzAlbumType, &a.MbzAlbumComment, &a.MbzReleaseGroupID, &a.FolderIds, &a.ExplicitStatus,
			&a.RgAlbumGain, &a.RgAlbumPeak, &a.Description, &a.SmallImageUrl, &a.MediumImageUrl,
			&a.LargeImageUrl, &a.ExternalUrl, &a.ExternalInfoUpdatedAt, &a.Genre, &a.Tags, &a.Missing,
			&a.ImportedAt, &a.CreatedAt, &a.UpdatedAt,
		); err == nil {
			albums = append(albums, *toModelAlbum(a))
		}
	}
	return albums, nil
}

// REST Repository integration
func (r *albumRepository) Count(_ ...rest.QueryOptions) (int64, error) {
	return r.CountAll()
}

func (r *albumRepository) Read(id string) (any, error) {
	return r.Get(id)
}

func (r *albumRepository) ReadAll(options ...rest.QueryOptions) (any, error) {
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

func (r *albumRepository) EntityName() string {
	return "album"
}

func (r *albumRepository) NewInstance() any {
	return &model.Album{}
}

var _ model.AlbumRepository = (*albumRepository)(nil)
var _ rest.Repository = (*albumRepository)(nil)

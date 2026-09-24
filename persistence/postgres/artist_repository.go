package postgres

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/deluan/rest"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/model/id"
	"github.com/navidrome/navidrome/model/request"
	"github.com/navidrome/navidrome/persistence/postgres/pgdb"
)

type artistRepository struct {
	ctx     context.Context
	queries *pgdb.Queries
	pool    *pgxpool.Pool
}

func NewArtistRepository(ctx context.Context, q *pgdb.Queries, p *pgxpool.Pool) model.ArtistRepository {
	return &artistRepository{
		ctx:     ctx,
		queries: q,
		pool:    p,
	}
}

func toModelArtist(a pgdb.Artist) *model.Artist {
	res := &model.Artist{
		ID:              a.ID,
		Name:            a.Name,
		SortArtistName:  a.SortArtistName,
		OrderArtistName: a.OrderArtistName,
		MbzArtistID:     a.MbzArtistID,
		Biography:       a.Biography,
		SmallImageUrl:   a.SmallImageUrl,
		MediumImageUrl:  a.MediumImageUrl,
		LargeImageUrl:   a.LargeImageUrl,
		ExternalUrl:     a.ExternalUrl,
		Missing:         a.Missing,
		UploadedImage:   a.UploadedImage,
	}
	if a.ExternalInfoUpdatedAt.Valid {
		t := a.ExternalInfoUpdatedAt.Time
		res.ExternalInfoUpdatedAt = &t
	}
	if a.CreatedAt.Valid {
		t := a.CreatedAt.Time
		res.CreatedAt = &t
	}
	if a.UpdatedAt.Valid {
		t := a.UpdatedAt.Time
		res.UpdatedAt = &t
	}
	return res
}

func (r *artistRepository) Get(id string) (*model.Artist, error) {
	a, err := r.queries.GetArtistByID(r.ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return toModelArtist(a), nil
}

func (r *artistRepository) GetAll(options ...model.QueryOptions) (model.Artists, error) {
	limit := int32(1000)
	offset := int32(0)
	orderCol := "sort_artist_name"
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
			case "created_at":
				orderCol = "created_at"
			case "name":
				orderCol = "name"
			default:
				orderCol = "sort_artist_name"
			}
			if strings.ToUpper(opt.Order) == "DESC" {
				orderDir = "DESC"
			}
		}
	}

	query := fmt.Sprintf(`
		SELECT id, name, sort_artist_name, order_artist_name, mbz_artist_id, biography,
		       small_image_url, medium_image_url, large_image_url, external_url,
		       similar_artists, external_info_updated_at, missing, uploaded_image,
		       created_at, updated_at
		FROM artist
		WHERE missing = FALSE
		ORDER BY %s %s, name ASC
		LIMIT $1 OFFSET $2
	`, orderCol, orderDir)

	rows, err := r.pool.Query(r.ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artists model.Artists
	for rows.Next() {
		var a pgdb.Artist
		if err := rows.Scan(
			&a.ID, &a.Name, &a.SortArtistName, &a.OrderArtistName, &a.MbzArtistID, &a.Biography,
			&a.SmallImageUrl, &a.MediumImageUrl, &a.LargeImageUrl, &a.ExternalUrl,
			&a.SimilarArtists, &a.ExternalInfoUpdatedAt, &a.Missing, &a.UploadedImage,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		artists = append(artists, *toModelArtist(a))
	}
	return artists, nil
}

func (r *artistRepository) CountAll(_ ...model.QueryOptions) (int64, error) {
	return r.queries.CountArtists(r.ctx)
}

func (r *artistRepository) Exists(id string) (bool, error) {
	_, err := r.Get(id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *artistRepository) Put(a *model.Artist, _ ...string) error {
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	params := pgdb.UpsertArtistParams{
		ID:              a.ID,
		Name:            a.Name,
		SortArtistName:  a.SortArtistName,
		OrderArtistName: a.OrderArtistName,
		MbzArtistID:     a.MbzArtistID,
		Biography:       a.Biography,
		SmallImageUrl:   a.SmallImageUrl,
		MediumImageUrl:  a.MediumImageUrl,
		LargeImageUrl:   a.LargeImageUrl,
		ExternalUrl:     a.ExternalUrl,
		SimilarArtists:  "",
		Missing:         a.Missing,
		UploadedImage:   a.UploadedImage,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	_, err := r.queries.UpsertArtist(r.ctx, params)
	return err
}

func (r *artistRepository) UpdateExternalInfo(a *model.Artist) error {
	return r.Put(a)
}

func (r *artistRepository) GetCursor(options ...model.QueryOptions) (model.ArtistCursor, error) {
	artists, err := r.GetAll(options...)
	if err != nil {
		return nil, err
	}
	return func(yield func(model.Artist, error) bool) {
		for _, a := range artists {
			if !yield(a, nil) {
				return
			}
		}
	}, nil
}

func getArtistIndexKey(a model.Artist) string {
	name := strings.TrimSpace(a.SortArtistName)
	if name == "" {
		name = strings.TrimSpace(a.Name)
	}
	if name == "" {
		return "#"
	}
	r := unicode.ToUpper([]rune(name)[0])
	if r >= 'A' && r <= 'Z' {
		return string(r)
	}
	return "#"
}

func (r *artistRepository) GetIndex(_ bool, _ []int, _ ...model.Role) (model.ArtistIndexes, error) {
	artists, err := r.GetAll(model.QueryOptions{Sort: "name"})
	if err != nil {
		return nil, err
	}

	groups := make(map[string]model.Artists)
	for _, a := range artists {
		k := getArtistIndexKey(a)
		groups[k] = append(groups[k], a)
	}

	var result model.ArtistIndexes
	for k, v := range groups {
		result = append(result, model.ArtistIndex{ID: k, Artists: v})
	}
	slices.SortFunc(result, func(a, b model.ArtistIndex) int {
		return cmp.Compare(a.ID, b.ID)
	})
	return result, nil
}

func (r *artistRepository) RefreshPlayCounts() (int64, error) {
	return 0, nil
}

func (r *artistRepository) RefreshStats(_ bool) (int64, error) {
	return 0, nil
}

func (r *artistRepository) IncPlayCount(itemID string, ts time.Time) error {
	user, ok := request.UserFrom(r.ctx)
	if !ok {
		return nil
	}
	annID := id.NewHash(user.ID + itemID + "artist")
	_, err := r.pool.Exec(r.ctx, `
		INSERT INTO annotation (ann_id, user_id, item_id, item_type, play_count, play_date)
		VALUES ($1, $2, $3, 'artist', 1, $4)
		ON CONFLICT (user_id, item_id, item_type)
		DO UPDATE SET play_count = annotation.play_count + 1, play_date = EXCLUDED.play_date
	`, annID, user.ID, itemID, ts)
	return err
}

func (r *artistRepository) SetStar(starred bool, itemIDs ...string) error {
	user, ok := request.UserFrom(r.ctx)
	if !ok {
		return nil
	}
	now := time.Now()
	for _, itemID := range itemIDs {
		annID := id.NewHash(user.ID + itemID + "artist")
		_, _ = r.pool.Exec(r.ctx, `
			INSERT INTO annotation (ann_id, user_id, item_id, item_type, starred, starred_at)
			VALUES ($1, $2, $3, 'artist', $4, $5)
			ON CONFLICT (user_id, item_id, item_type)
			DO UPDATE SET starred = EXCLUDED.starred, starred_at = EXCLUDED.starred_at
		`, annID, user.ID, itemID, starred, now)
	}
	return nil
}

func (r *artistRepository) SetRating(rating int, itemID string) error {
	user, ok := request.UserFrom(r.ctx)
	if !ok {
		return nil
	}
	now := time.Now()
	annID := id.NewHash(user.ID + itemID + "artist")
	_, err := r.pool.Exec(r.ctx, `
		INSERT INTO annotation (ann_id, user_id, item_id, item_type, rating, rated_at)
		VALUES ($1, $2, $3, 'artist', $4, $5)
		ON CONFLICT (user_id, item_id, item_type)
		DO UPDATE SET rating = EXCLUDED.rating, rated_at = EXCLUDED.rated_at
	`, annID, user.ID, itemID, rating, now)
	return err
}

func (r *artistRepository) ReassignAnnotation(prevID string, newID string) error {
	_, err := r.pool.Exec(r.ctx, `
		UPDATE annotation SET item_id = $1 WHERE item_id = $2 AND item_type = 'artist'
	`, newID, prevID)
	return err
}

func (r *artistRepository) Search(q string, options ...model.QueryOptions) (model.Artists, error) {
	limit := int32(50)
	if len(options) > 0 && options[0].Max > 0 {
		limit = int32(options[0].Max)
	}
	pattern := "%" + q + "%"
	query := `
		SELECT id, name, sort_artist_name, order_artist_name, mbz_artist_id, biography,
		       small_image_url, medium_image_url, large_image_url, external_url,
		       similar_artists, external_info_updated_at, missing, uploaded_image,
		       created_at, updated_at
		FROM artist
		WHERE missing = FALSE AND name ILIKE $1
		ORDER BY sort_artist_name ASC
		LIMIT $2
	`
	rows, err := r.pool.Query(r.ctx, query, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artists model.Artists
	for rows.Next() {
		var a pgdb.Artist
		if err := rows.Scan(
			&a.ID, &a.Name, &a.SortArtistName, &a.OrderArtistName, &a.MbzArtistID, &a.Biography,
			&a.SmallImageUrl, &a.MediumImageUrl, &a.LargeImageUrl, &a.ExternalUrl,
			&a.SimilarArtists, &a.ExternalInfoUpdatedAt, &a.Missing, &a.UploadedImage,
			&a.CreatedAt, &a.UpdatedAt,
		); err == nil {
			artists = append(artists, *toModelArtist(a))
		}
	}
	return artists, nil
}

// REST Repository integration
func (r *artistRepository) Count(_ ...rest.QueryOptions) (int64, error) {
	return r.CountAll()
}

func (r *artistRepository) Read(id string) (any, error) {
	return r.Get(id)
}

func (r *artistRepository) ReadAll(options ...rest.QueryOptions) (any, error) {
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

func (r *artistRepository) EntityName() string {
	return "artist"
}

func (r *artistRepository) NewInstance() any {
	return &model.Artist{}
}

var _ model.ArtistRepository = (*artistRepository)(nil)
var _ rest.Repository = (*artistRepository)(nil)

package postgres

import (
	"context"
	"errors"
	"strconv"
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

type playlistRepository struct {
	ctx     context.Context
	queries *pgdb.Queries
	pool    *pgxpool.Pool
}

func NewPlaylistRepository(ctx context.Context, q *pgdb.Queries, p *pgxpool.Pool) model.PlaylistRepository {
	return &playlistRepository{
		ctx:     ctx,
		queries: q,
		pool:    p,
	}
}

func toModelPlaylist(p pgdb.Playlist) *model.Playlist {
	return &model.Playlist{
		ID:               p.ID,
		Name:             p.Name,
		Comment:          p.Comment,
		Duration:         p.Duration,
		Size:             p.Size,
		SongCount:        int(p.SongCount),
		OwnerID:          p.OwnerID,
		Public:           p.Public,
		Path:             p.Path,
		Sync:             p.Sync,
		UploadedImage:    p.UploadedImage,
		ExternalImageURL: p.ExternalImageUrl,
		ImportedHash:     p.ImportedHash,
		CreatedAt:        p.CreatedAt.Time,
		UpdatedAt:        p.UpdatedAt.Time,
	}
}

func (r *playlistRepository) Get(id string) (*model.Playlist, error) {
	p, err := r.queries.GetPlaylistByID(r.ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return toModelPlaylist(p), nil
}

func (r *playlistRepository) GetWithTracks(id string, _, _ bool) (*model.Playlist, error) {
	p, err := r.Get(id)
	if err != nil {
		return nil, err
	}
	dbTracks, err := r.queries.GetPlaylistTracks(r.ctx, id)
	if err == nil {
		tracks := make(model.PlaylistTracks, len(dbTracks))
		for i, mf := range dbTracks {
			tracks[i] = model.PlaylistTrack{
				ID:          strconv.Itoa(i + 1),
				PlaylistID:  id,
				MediaFileID: mf.ID,
				MediaFile:   *toModelMediaFile(mf),
			}
		}
		p.SetTracks(tracks)
	}
	return p, nil
}

func (r *playlistRepository) GetAll(_ ...model.QueryOptions) (model.Playlists, error) {
	ownerID := ""
	if user, ok := request.UserFrom(r.ctx); ok {
		ownerID = user.ID
	}

	dbPlaylists, err := r.queries.ListPlaylists(r.ctx, ownerID)
	if err != nil {
		return nil, err
	}

	playlists := make(model.Playlists, len(dbPlaylists))
	for i, p := range dbPlaylists {
		playlists[i] = *toModelPlaylist(p)
	}
	return playlists, nil
}

func (r *playlistRepository) CountAll(_ ...model.QueryOptions) (int64, error) {
	var count int64
	err := r.pool.QueryRow(r.ctx, `SELECT count(*) FROM playlist`).Scan(&count)
	return count, err
}

func (r *playlistRepository) Exists(id string) (bool, error) {
	_, err := r.Get(id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *playlistRepository) Put(pls *model.Playlist, _ ...string) error {
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	params := pgdb.UpsertPlaylistParams{
		ID:               pls.ID,
		Name:             pls.Name,
		Comment:          pls.Comment,
		Duration:         pls.Duration,
		Size:             pls.Size,
		SongCount:        int32(pls.SongCount),
		OwnerID:          pls.OwnerID,
		Public:           pls.Public,
		Path:             pls.Path,
		Sync:             pls.Sync,
		UploadedImage:    pls.UploadedImage,
		ExternalImageUrl: pls.ExternalImageURL,
		ImportedHash:     pls.ImportedHash,
		Rules:            nil,
		EvaluatedAt:      pgtype.Timestamptz{Valid: false},
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	_, err := r.queries.UpsertPlaylist(r.ctx, params)
	return err
}

func (r *playlistRepository) Delete(id string) error {
	return r.queries.DeletePlaylist(r.ctx, id)
}

func (r *playlistRepository) FindByPath(_ string) (*model.Playlist, error) {
	return nil, model.ErrNotFound
}

func (r *playlistRepository) GetCursor(options ...model.QueryOptions) (model.PlaylistCursor, error) {
	pls, err := r.GetAll(options...)
	if err != nil {
		return nil, err
	}
	return func(yield func(model.Playlist, error) bool) {
		for _, p := range pls {
			if !yield(p, nil) {
				return
			}
		}
	}, nil
}

func (r *playlistRepository) GetPlaylists(mediaFileId string) (model.Playlists, error) {
	rows, err := r.pool.Query(r.ctx, `
		SELECT p.id, p.name, p.comment, p.duration, p.size, p.song_count,
		       p.owner_id, p.public, p.path, p.sync, p.uploaded_image,
		       p.external_image_url, p.imported_hash, p.rules, p.evaluated_at,
		       p.created_at, p.updated_at
		FROM playlist p
		JOIN playlist_tracks pt ON p.id = pt.playlist_id
		WHERE pt.media_file_id = $1
		ORDER BY p.name ASC
	`, mediaFileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists model.Playlists
	for rows.Next() {
		var p pgdb.Playlist
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Comment, &p.Duration, &p.Size, &p.SongCount,
			&p.OwnerID, &p.Public, &p.Path, &p.Sync, &p.UploadedImage,
			&p.ExternalImageUrl, &p.ImportedHash, &p.Rules, &p.EvaluatedAt,
			&p.CreatedAt, &p.UpdatedAt,
		); err == nil {
			playlists = append(playlists, *toModelPlaylist(p))
		}
	}
	return playlists, nil
}

func (r *playlistRepository) Tracks(playlistId string, _ bool) model.PlaylistTrackRepository {
	return newPlaylistTrackRepository(r.ctx, playlistId, r.queries, r.pool)
}

func (r *playlistRepository) IncPlayCount(itemID string, ts time.Time) error {
	user, ok := request.UserFrom(r.ctx)
	if !ok {
		return nil
	}
	annID := id.NewHash(user.ID + itemID + "playlist")
	_, err := r.pool.Exec(r.ctx, `
		INSERT INTO annotation (ann_id, user_id, item_id, item_type, play_count, play_date)
		VALUES ($1, $2, $3, 'playlist', 1, $4)
		ON CONFLICT (user_id, item_id, item_type)
		DO UPDATE SET play_count = annotation.play_count + 1, play_date = EXCLUDED.play_date
	`, annID, user.ID, itemID, ts)
	return err
}

func (r *playlistRepository) SetStar(starred bool, itemIDs ...string) error {
	user, ok := request.UserFrom(r.ctx)
	if !ok {
		return nil
	}
	now := time.Now()
	for _, itemID := range itemIDs {
		annID := id.NewHash(user.ID + itemID + "playlist")
		_, _ = r.pool.Exec(r.ctx, `
			INSERT INTO annotation (ann_id, user_id, item_id, item_type, starred, starred_at)
			VALUES ($1, $2, $3, 'playlist', $4, $5)
			ON CONFLICT (user_id, item_id, item_type)
			DO UPDATE SET starred = EXCLUDED.starred, starred_at = EXCLUDED.starred_at
		`, annID, user.ID, itemID, starred, now)
	}
	return nil
}

func (r *playlistRepository) SetRating(rating int, itemID string) error {
	user, ok := request.UserFrom(r.ctx)
	if !ok {
		return nil
	}
	now := time.Now()
	annID := id.NewHash(user.ID + itemID + "playlist")
	_, err := r.pool.Exec(r.ctx, `
		INSERT INTO annotation (ann_id, user_id, item_id, item_type, rating, rated_at)
		VALUES ($1, $2, $3, 'playlist', $4, $5)
		ON CONFLICT (user_id, item_id, item_type)
		DO UPDATE SET rating = EXCLUDED.rating, rated_at = EXCLUDED.rated_at
	`, annID, user.ID, itemID, rating, now)
	return err
}

func (r *playlistRepository) ReassignAnnotation(prevID string, newID string) error {
	_, err := r.pool.Exec(r.ctx, `
		UPDATE annotation SET item_id = $1 WHERE item_id = $2 AND item_type = 'playlist'
	`, newID, prevID)
	return err
}

// REST Repository integration
func (r *playlistRepository) Count(_ ...rest.QueryOptions) (int64, error) {
	return r.CountAll()
}

func (r *playlistRepository) Read(id string) (any, error) {
	return r.GetWithTracks(id, false, false)
}

func (r *playlistRepository) ReadAll(options ...rest.QueryOptions) (any, error) {
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

func (r *playlistRepository) EntityName() string {
	return "playlist"
}

func (r *playlistRepository) NewInstance() any {
	return &model.Playlist{}
}

func (r *playlistRepository) Save(entity any) (string, error) {
	p, ok := entity.(*model.Playlist)
	if !ok {
		return "", errors.New("invalid playlist entity")
	}
	if p.ID == "" {
		p.ID = id.NewRandom()
	}
	user, ok := request.UserFrom(r.ctx)
	if ok && p.OwnerID == "" {
		p.OwnerID = user.ID
	}
	if err := r.Put(p); err != nil {
		return "", err
	}
	return p.ID, nil
}

func (r *playlistRepository) Update(id string, entity any, cols ...string) error {
	p, ok := entity.(*model.Playlist)
	if !ok {
		return errors.New("invalid playlist entity")
	}
	p.ID = id
	return r.Put(p, cols...)
}

var _ model.PlaylistRepository = (*playlistRepository)(nil)
var _ rest.Repository = (*playlistRepository)(nil)
var _ rest.Persistable = (*playlistRepository)(nil)

// --- PlaylistTrackRepository ---

type playlistTrackRepository struct {
	ctx        context.Context
	playlistID string
	queries    *pgdb.Queries
	pool       *pgxpool.Pool
}

func newPlaylistTrackRepository(ctx context.Context, playlistID string, q *pgdb.Queries, p *pgxpool.Pool) model.PlaylistTrackRepository {
	return &playlistTrackRepository{
		ctx:        ctx,
		playlistID: playlistID,
		queries:    q,
		pool:       p,
	}
}

func (r *playlistTrackRepository) CountAll(_ ...model.QueryOptions) (int64, error) {
	var count int64
	err := r.pool.QueryRow(r.ctx, `SELECT count(*) FROM playlist_tracks WHERE playlist_id = $1`, r.playlistID).Scan(&count)
	return count, err
}

func (r *playlistTrackRepository) GetAll(_ ...model.QueryOptions) (model.PlaylistTracks, error) {
	dbTracks, err := r.queries.GetPlaylistTracks(r.ctx, r.playlistID)
	if err != nil {
		return nil, err
	}
	tracks := make(model.PlaylistTracks, len(dbTracks))
	for i, mf := range dbTracks {
		tracks[i] = model.PlaylistTrack{
			ID:          strconv.Itoa(i + 1),
			PlaylistID:  r.playlistID,
			MediaFileID: mf.ID,
			MediaFile:   *toModelMediaFile(mf),
		}
	}
	return tracks, nil
}

func (r *playlistTrackRepository) GetCursor(options ...model.QueryOptions) (model.PlaylistTrackCursor, error) {
	tracks, err := r.GetAll(options...)
	if err != nil {
		return nil, err
	}
	return func(yield func(model.PlaylistTrack, error) bool) {
		for _, t := range tracks {
			if !yield(t, nil) {
				return
			}
		}
	}, nil
}

func (r *playlistTrackRepository) GetAlbumIDs(_ ...model.QueryOptions) ([]string, error) {
	rows, err := r.pool.Query(r.ctx, `
		SELECT DISTINCT mf.album_id
		FROM media_file mf
		JOIN playlist_tracks pt ON mf.id = pt.media_file_id
		WHERE pt.playlist_id = $1 AND mf.album_id != ''
	`, r.playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albumIDs []string
	for rows.Next() {
		var aid string
		if err := rows.Scan(&aid); err == nil {
			albumIDs = append(albumIDs, aid)
		}
	}
	return albumIDs, nil
}

func (r *playlistTrackRepository) GetMediaFileIDs(_ ...model.QueryOptions) ([]string, error) {
	rows, err := r.pool.Query(r.ctx, `
		SELECT media_file_id
		FROM playlist_tracks
		WHERE playlist_id = $1
		ORDER BY id ASC
	`, r.playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mfIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			mfIDs = append(mfIDs, id)
		}
	}
	return mfIDs, nil
}

func (r *playlistTrackRepository) Add(mediaFileIds []string) (int, error) {
	var maxID int32
	_ = r.pool.QueryRow(r.ctx, `SELECT COALESCE(MAX(id), 0) FROM playlist_tracks WHERE playlist_id = $1`, r.playlistID).Scan(&maxID)

	added := 0
	for _, mfId := range mediaFileIds {
		maxID++
		err := r.queries.AddPlaylistTrack(r.ctx, pgdb.AddPlaylistTrackParams{
			ID:          maxID,
			PlaylistID:  r.playlistID,
			MediaFileID: mfId,
		})
		if err == nil {
			added++
		}
	}
	return added, nil
}

func (r *playlistTrackRepository) Insert(mediaFileIds []string, _ int) (int, error) {
	return r.Add(mediaFileIds)
}

func (r *playlistTrackRepository) AddAlbums(albumIds []string) (int, error) {
	var mfIDs []string
	for _, aID := range albumIds {
		rows, err := r.pool.Query(r.ctx, `SELECT id FROM media_file WHERE album_id = $1 AND missing = FALSE ORDER BY disc_number, track_number`, aID)
		if err == nil {
			for rows.Next() {
				var id string
				if err := rows.Scan(&id); err == nil {
					mfIDs = append(mfIDs, id)
				}
			}
			rows.Close()
		}
	}
	return r.Add(mfIDs)
}

func (r *playlistTrackRepository) AddArtists(artistIds []string) (int, error) {
	var mfIDs []string
	for _, artID := range artistIds {
		rows, err := r.pool.Query(r.ctx, `SELECT id FROM media_file WHERE artist_id = $1 AND missing = FALSE ORDER BY disc_number, track_number`, artID)
		if err == nil {
			for rows.Next() {
				var id string
				if err := rows.Scan(&id); err == nil {
					mfIDs = append(mfIDs, id)
				}
			}
			rows.Close()
		}
	}
	return r.Add(mfIDs)
}

func (r *playlistTrackRepository) AddDiscs(discs []model.DiscID) (int, error) {
	var mfIDs []string
	for _, d := range discs {
		rows, err := r.pool.Query(r.ctx, `SELECT id FROM media_file WHERE album_id = $1 AND disc_number = $2 AND missing = FALSE ORDER BY track_number`, d.AlbumID, d.DiscNumber)
		if err == nil {
			for rows.Next() {
				var id string
				if err := rows.Scan(&id); err == nil {
					mfIDs = append(mfIDs, id)
				}
			}
			rows.Close()
		}
	}
	return r.Add(mfIDs)
}

func (r *playlistTrackRepository) Delete(ids ...string) error {
	for _, idStr := range ids {
		if idInt, err := strconv.Atoi(idStr); err == nil {
			_, _ = r.pool.Exec(r.ctx, `DELETE FROM playlist_tracks WHERE playlist_id = $1 AND id = $2`, r.playlistID, idInt)
		}
	}
	return nil
}

func (r *playlistTrackRepository) DeleteAll() error {
	return r.queries.ClearPlaylistTracks(r.ctx, r.playlistID)
}

func (r *playlistTrackRepository) Reorder(_ int, _ int) error {
	return nil
}

func (r *playlistTrackRepository) Count(_ ...rest.QueryOptions) (int64, error) {
	return r.CountAll()
}

func (r *playlistTrackRepository) Read(id string) (any, error) {
	trackID, err := strconv.Atoi(id)
	if err != nil {
		return nil, model.ErrNotFound
	}
	var mfID string
	err = r.pool.QueryRow(r.ctx, `SELECT media_file_id FROM playlist_tracks WHERE playlist_id = $1 AND id = $2`, r.playlistID, trackID).Scan(&mfID)
	if err != nil {
		return nil, model.ErrNotFound
	}
	mf, err := r.queries.GetMediaFileByID(r.ctx, mfID)
	if err != nil {
		return nil, err
	}
	return &model.PlaylistTrack{
		ID:          id,
		PlaylistID:  r.playlistID,
		MediaFileID: mfID,
		MediaFile:   *toModelMediaFile(mf),
	}, nil
}

func (r *playlistTrackRepository) ReadAll(options ...rest.QueryOptions) (any, error) {
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

func (r *playlistTrackRepository) EntityName() string {
	return "playlist_track"
}

func (r *playlistTrackRepository) NewInstance() any {
	return &model.PlaylistTrack{}
}

var _ model.PlaylistTrackRepository = (*playlistTrackRepository)(nil)
var _ rest.Repository = (*playlistTrackRepository)(nil)

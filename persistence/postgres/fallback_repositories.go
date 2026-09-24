package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/deluan/rest"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/persistence/postgres/pgdb"
)

// --- ScrobbleBufferRepository ---

type scrobbleBufferRepository struct{}

func NewScrobbleBufferRepository(_ context.Context) model.ScrobbleBufferRepository {
	return &scrobbleBufferRepository{}
}

func (r *scrobbleBufferRepository) UserIDs(_ string) ([]string, error) {
	return []string{}, nil
}

func (r *scrobbleBufferRepository) Enqueue(_, _, _ string, _ time.Time) error {
	return nil
}

func (r *scrobbleBufferRepository) Next(_, _ string) (*model.ScrobbleEntry, error) {
	return nil, model.ErrNotFound
}

func (r *scrobbleBufferRepository) Dequeue(_ *model.ScrobbleEntry) error {
	return nil
}

func (r *scrobbleBufferRepository) Length() (int64, error) {
	return 0, nil
}

func (r *scrobbleBufferRepository) Discard(_ string) error {
	return nil
}

var _ model.ScrobbleBufferRepository = (*scrobbleBufferRepository)(nil)

// --- TranscodingRepository ---

type transcodingRepository struct{}

func NewTranscodingRepository(_ context.Context) model.TranscodingRepository {
	return &transcodingRepository{}
}

func (r *transcodingRepository) Get(_ string) (*model.Transcoding, error) {
	return nil, model.ErrNotFound
}

func (r *transcodingRepository) CountAll(_ ...model.QueryOptions) (int64, error) {
	return 0, nil
}

func (r *transcodingRepository) Put(_ *model.Transcoding) error {
	return nil
}

func (r *transcodingRepository) FindByFormat(_ string) (*model.Transcoding, error) {
	return nil, model.ErrNotFound
}

func (r *transcodingRepository) Count(_ ...rest.QueryOptions) (int64, error) {
	return 0, nil
}

func (r *transcodingRepository) Read(_ string) (any, error) {
	return nil, model.ErrNotFound
}

func (r *transcodingRepository) ReadAll(_ ...rest.QueryOptions) (any, error) {
	return model.Transcodings{}, nil
}

func (r *transcodingRepository) EntityName() string {
	return "transcoding"
}

func (r *transcodingRepository) NewInstance() any {
	return &model.Transcoding{}
}

func (r *transcodingRepository) Save(_ any) (string, error) {
	return "", nil
}

func (r *transcodingRepository) Update(_ string, _ any, _ ...string) error {
	return nil
}

func (r *transcodingRepository) Delete(_ string) error {
	return nil
}

var _ model.TranscodingRepository = (*transcodingRepository)(nil)
var _ rest.Repository = (*transcodingRepository)(nil)
var _ rest.Persistable = (*transcodingRepository)(nil)

// --- PlayQueueRepository ---

type playQueueRepository struct{}

func NewPlayQueueRepository(_ context.Context, _ *pgxpool.Pool) model.PlayQueueRepository {
	return &playQueueRepository{}
}

func (r *playQueueRepository) Store(_ *model.PlayQueue, _ ...string) error {
	return nil
}

func (r *playQueueRepository) Retrieve(userId string) (*model.PlayQueue, error) {
	return &model.PlayQueue{UserID: userId, Items: model.MediaFiles{}}, nil
}

func (r *playQueueRepository) RetrieveWithMediaFiles(userId string) (*model.PlayQueue, error) {
	return &model.PlayQueue{UserID: userId, Items: model.MediaFiles{}}, nil
}

func (r *playQueueRepository) Clear(_ string) error {
	return nil
}

var _ model.PlayQueueRepository = (*playQueueRepository)(nil)

// --- UserPropsRepository ---

type userPropsRepository struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

func NewUserPropsRepository(ctx context.Context, pool *pgxpool.Pool) model.UserPropsRepository {
	return &userPropsRepository{ctx: ctx, pool: pool}
}

func (r *userPropsRepository) Put(userId, key, value string) error {
	_, err := r.pool.Exec(r.ctx, `
		INSERT INTO user_props (user_id, key, value)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, key) DO UPDATE SET value = EXCLUDED.value
	`, userId, key, value)
	return err
}

func (r *userPropsRepository) Get(userId, key string) (string, error) {
	var val string
	err := r.pool.QueryRow(r.ctx, `SELECT value FROM user_props WHERE user_id = $1 AND key = $2`, userId, key).Scan(&val)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", model.ErrNotFound
		}
		return "", err
	}
	return val, nil
}

func (r *userPropsRepository) Delete(userId, key string) error {
	_, err := r.pool.Exec(r.ctx, `DELETE FROM user_props WHERE user_id = $1 AND key = $2`, userId, key)
	return err
}

func (r *userPropsRepository) DefaultGet(userId, key, defaultValue string) (string, error) {
	val, err := r.Get(userId, key)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return defaultValue, nil
		}
		return "", err
	}
	return val, nil
}

var _ model.UserPropsRepository = (*userPropsRepository)(nil)

// --- PlayerRepository ---

type playerRepository struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

func NewPlayerRepository(ctx context.Context, pool *pgxpool.Pool) model.PlayerRepository {
	return &playerRepository{ctx: ctx, pool: pool}
}

func (r *playerRepository) Get(_ string) (*model.Player, error) {
	return nil, model.ErrNotFound
}

func (r *playerRepository) FindMatch(_, _, _ string) (*model.Player, error) {
	return nil, model.ErrNotFound
}

func (r *playerRepository) Put(_ *model.Player) error {
	return nil
}

func (r *playerRepository) CountAll(_ ...model.QueryOptions) (int64, error) {
	return 0, nil
}

func (r *playerRepository) CountByClient(_ ...model.QueryOptions) (map[string]int64, error) {
	return map[string]int64{}, nil
}

func (r *playerRepository) Count(_ ...rest.QueryOptions) (int64, error) {
	return 0, nil
}

func (r *playerRepository) Read(_ string) (any, error) {
	return nil, model.ErrNotFound
}

func (r *playerRepository) ReadAll(_ ...rest.QueryOptions) (any, error) {
	return model.Players{}, nil
}

func (r *playerRepository) EntityName() string {
	return "player"
}

func (r *playerRepository) NewInstance() any {
	return &model.Player{}
}

func (r *playerRepository) Save(_ any) (string, error) {
	return "", nil
}

func (r *playerRepository) Update(_ string, _ any, _ ...string) error {
	return nil
}

func (r *playerRepository) Delete(_ string) error {
	return nil
}

var _ model.PlayerRepository = (*playerRepository)(nil)
var _ rest.Repository = (*playerRepository)(nil)
var _ rest.Persistable = (*playerRepository)(nil)

// --- RadioRepository ---

type radioRepository struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

func NewRadioRepository(ctx context.Context, pool *pgxpool.Pool) model.RadioRepository {
	return &radioRepository{ctx: ctx, pool: pool}
}

func (r *radioRepository) CountAll(_ ...model.QueryOptions) (int64, error) {
	return 0, nil
}

func (r *radioRepository) Delete(_ string) error {
	return nil
}

func (r *radioRepository) Exists(_ string) (bool, error) {
	return false, nil
}

func (r *radioRepository) Get(_ string) (*model.Radio, error) {
	return nil, model.ErrNotFound
}

func (r *radioRepository) GetAll(_ ...model.QueryOptions) (model.Radios, error) {
	return model.Radios{}, nil
}

func (r *radioRepository) Put(_ *model.Radio, _ ...string) error {
	return nil
}

func (r *radioRepository) Count(_ ...rest.QueryOptions) (int64, error) {
	return 0, nil
}

func (r *radioRepository) Read(_ string) (any, error) {
	return nil, model.ErrNotFound
}

func (r *radioRepository) ReadAll(_ ...rest.QueryOptions) (any, error) {
	return model.Radios{}, nil
}

func (r *radioRepository) EntityName() string {
	return "radio"
}

func (r *radioRepository) NewInstance() any {
	return &model.Radio{}
}

func (r *radioRepository) Save(_ any) (string, error) {
	return "", nil
}

func (r *radioRepository) Update(_ string, _ any, _ ...string) error {
	return nil
}

var _ model.RadioRepository = (*radioRepository)(nil)
var _ rest.Repository = (*radioRepository)(nil)
var _ rest.Persistable = (*radioRepository)(nil)

// --- ShareRepository ---

type shareRepository struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

func NewShareRepository(ctx context.Context, pool *pgxpool.Pool) model.ShareRepository {
	return &shareRepository{ctx: ctx, pool: pool}
}

func (r *shareRepository) Exists(_ string) (bool, error) {
	return false, nil
}

func (r *shareRepository) Get(_ string) (*model.Share, error) {
	return nil, model.ErrNotFound
}

func (r *shareRepository) GetWithResources(_ string) (*model.Share, error) {
	return nil, model.ErrNotFound
}

func (r *shareRepository) CountAll(_ ...model.QueryOptions) (int64, error) {
	return 0, nil
}

func (r *shareRepository) GetAll(_ ...model.QueryOptions) (model.Shares, error) {
	return model.Shares{}, nil
}

func (r *shareRepository) Put(_ *model.Share) error {
	return nil
}

func (r *shareRepository) Delete(_ string) error {
	return nil
}

func (r *shareRepository) Count(_ ...rest.QueryOptions) (int64, error) {
	return 0, nil
}

func (r *shareRepository) Read(_ string) (any, error) {
	return nil, model.ErrNotFound
}

func (r *shareRepository) ReadAll(_ ...rest.QueryOptions) (any, error) {
	return model.Shares{}, nil
}

func (r *shareRepository) EntityName() string {
	return "share"
}

func (r *shareRepository) NewInstance() any {
	return &model.Share{}
}

func (r *shareRepository) Save(_ any) (string, error) {
	return "", nil
}

func (r *shareRepository) Update(_ string, _ any, _ ...string) error {
	return nil
}

var _ model.ShareRepository = (*shareRepository)(nil)
var _ rest.Repository = (*shareRepository)(nil)
var _ rest.Persistable = (*shareRepository)(nil)

// --- GenreRepository ---

type genreRepository struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

func NewGenreRepository(ctx context.Context, pool *pgxpool.Pool) model.GenreRepository {
	return &genreRepository{ctx: ctx, pool: pool}
}

func (r *genreRepository) GetAll(_ ...model.QueryOptions) (model.Genres, error) {
	rows, err := r.pool.Query(r.ctx, `
		SELECT genre, count(*) as album_count, sum(song_count) as song_count
		FROM album
		WHERE missing = FALSE AND genre != ''
		GROUP BY genre
		ORDER BY genre ASC
	`)
	if err != nil {
		return model.Genres{}, nil
	}
	defer rows.Close()

	var genres model.Genres
	for rows.Next() {
		var g model.Genre
		var albumCount int64
		var songCount int64
		if err := rows.Scan(&g.Name, &albumCount, &songCount); err == nil {
			g.ID = g.Name
			g.AlbumCount = int(albumCount)
			g.SongCount = int(songCount)
			genres = append(genres, g)
		}
	}
	return genres, nil
}

func (r *genreRepository) Get(id string) (*model.Genre, error) {
	return &model.Genre{ID: id, Name: id}, nil
}

func (r *genreRepository) Count(_ ...rest.QueryOptions) (int64, error) {
	var count int64
	_ = r.pool.QueryRow(r.ctx, `SELECT count(DISTINCT genre) FROM album WHERE missing = FALSE AND genre != ''`).Scan(&count)
	return count, nil
}

func (r *genreRepository) Read(id string) (any, error) {
	return r.Get(id)
}

func (r *genreRepository) ReadAll(_ ...rest.QueryOptions) (any, error) {
	return r.GetAll()
}

func (r *genreRepository) EntityName() string {
	return "genre"
}

func (r *genreRepository) NewInstance() any {
	return &model.Genre{}
}

var _ model.GenreRepository = (*genreRepository)(nil)
var _ rest.Repository = (*genreRepository)(nil)

// --- TagRepository ---

type tagRepository struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

func NewTagRepository(ctx context.Context, pool *pgxpool.Pool) model.TagRepository {
	return &tagRepository{ctx: ctx, pool: pool}
}

func (r *tagRepository) Add(_ int, _ ...model.Tag) error {
	return nil
}

func (r *tagRepository) UpdateCounts() error {
	return nil
}

func (r *tagRepository) GetAll(_ model.TagName, _ ...model.QueryOptions) (model.TagList, error) {
	return model.TagList{}, nil
}

func (r *tagRepository) Count(_ ...rest.QueryOptions) (int64, error) {
	return 0, nil
}

func (r *tagRepository) Read(_ string) (any, error) {
	return nil, model.ErrNotFound
}

func (r *tagRepository) ReadAll(_ ...rest.QueryOptions) (any, error) {
	return model.TagList{}, nil
}

func (r *tagRepository) EntityName() string {
	return "tag"
}

func (r *tagRepository) NewInstance() any {
	return &model.Tag{}
}

var _ model.TagRepository = (*tagRepository)(nil)
var _ rest.Repository = (*tagRepository)(nil)

// --- ArtworkRepository & ArtworkQueueRepository ---

type artworkRepository struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

func NewArtworkRepository(ctx context.Context, pool *pgxpool.Pool) model.ArtworkRepository {
	return &artworkRepository{ctx: ctx, pool: pool}
}

func (r *artworkRepository) GetImage(_ string) (*model.Artwork, error) {
	return nil, model.ErrNotFound
}

func (r *artworkRepository) PutImage(_ *model.Artwork) error {
	return nil
}

func (r *artworkRepository) PurgeOrphans(_ time.Time) (int64, error) {
	return 0, nil
}

func (r *artworkRepository) GetItemArtwork(_ model.Kind, _, _ string) (*model.ItemArtwork, error) {
	return nil, model.ErrNotFound
}

func (r *artworkRepository) PutItemArtwork(_ *model.ItemArtwork) error {
	return nil
}

func (r *artworkRepository) PutLastFailure(_ model.Kind, _, _, _ string) error {
	return nil
}

func (r *artworkRepository) DeleteForItems(_ model.Kind, _ []string) error {
	return nil
}

func (r *artworkRepository) GetInfoForItems(_ model.Kind, _ []string) (map[string]model.ItemArtworkInfo, error) {
	return map[string]model.ItemArtworkInfo{}, nil
}

func (r *artworkRepository) GetMimeByHash() (map[string]string, error) {
	return map[string]string{}, nil
}

func (r *artworkRepository) PurgeDanglingItems() (int64, error) {
	return 0, nil
}

var _ model.ArtworkRepository = (*artworkRepository)(nil)

type artworkQueueRepository struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

func NewArtworkQueueRepository(ctx context.Context, pool *pgxpool.Pool) model.ArtworkQueueRepository {
	return &artworkQueueRepository{ctx: ctx, pool: pool}
}

func (r *artworkQueueRepository) Get(_ model.Kind, _, _ string) (*model.ArtworkQueueItem, error) {
	return nil, model.ErrNotFound
}

func (r *artworkQueueRepository) Enqueue(_ ...model.ArtworkQueueItem) error {
	return nil
}

func (r *artworkQueueRepository) EnqueuePreservingBackoff(_ ...model.ArtworkQueueItem) error {
	return nil
}

func (r *artworkQueueRepository) EnqueueAllMissing(_ model.Kind, _ int) (int64, error) {
	return 0, nil
}

func (r *artworkQueueRepository) EnqueueIfMissing(_ ...model.ArtworkQueueItem) error {
	return nil
}

func (r *artworkQueueRepository) CountBySource(_ model.Kind, _ []string) (int64, error) {
	return 0, nil
}

func (r *artworkQueueRepository) SourcesInUse(_ model.Kind) ([]string, error) {
	return []string{}, nil
}

func (r *artworkQueueRepository) EnqueueBySource(_ model.Kind, _ []string, _ int) (int64, error) {
	return 0, nil
}

func (r *artworkQueueRepository) DequeueBatch(_ int, _ ...string) ([]model.ArtworkQueueItem, error) {
	return []model.ArtworkQueueItem{}, nil
}

func (r *artworkQueueRepository) MarkFailedIfUnchanged(_, _, _ string, _, _ time.Time, _ string) error {
	return nil
}

func (r *artworkQueueRepository) DeleteIfUnchanged(_, _, _ string, _ time.Time) error {
	return nil
}

func (r *artworkQueueRepository) Count() (int64, error) {
	return 0, nil
}

func (r *artworkQueueRepository) CountQueued(_ []model.Kind, _ []int) ([]model.ArtworkQueueStat, error) {
	return []model.ArtworkQueueStat{}, nil
}

func (r *artworkQueueRepository) PurgeDangling() (int64, error) {
	return 0, nil
}

func (r *artworkQueueRepository) PurgeQueued(_ []model.Kind, _ []int) (int64, error) {
	return 0, nil
}

var _ model.ArtworkQueueRepository = (*artworkQueueRepository)(nil)

// --- ScrobbleRepository ---

type scrobbleRepository struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

func NewScrobbleRepository(ctx context.Context, pool *pgxpool.Pool) model.ScrobbleRepository {
	return &scrobbleRepository{ctx: ctx, pool: pool}
}

func (r *scrobbleRepository) CountAll(_ ...model.QueryOptions) (int64, error) {
	return 0, nil
}

func (r *scrobbleRepository) Get(_ string) (*model.Scrobble, error) {
	return nil, model.ErrNotFound
}

func (r *scrobbleRepository) GetAll(_ ...model.QueryOptions) (model.Scrobbles, error) {
	return model.Scrobbles{}, nil
}

func (r *scrobbleRepository) RecordScrobble(_ string, _ time.Time) error {
	return nil
}

func (r *scrobbleRepository) Count(_ ...rest.QueryOptions) (int64, error) {
	return 0, nil
}

func (r *scrobbleRepository) Read(_ string) (any, error) {
	return nil, model.ErrNotFound
}

func (r *scrobbleRepository) ReadAll(_ ...rest.QueryOptions) (any, error) {
	return model.Scrobbles{}, nil
}

func (r *scrobbleRepository) EntityName() string {
	return "scrobble"
}

func (r *scrobbleRepository) NewInstance() any {
	return &model.Scrobble{}
}

var _ model.ScrobbleRepository = (*scrobbleRepository)(nil)
var _ rest.Repository = (*scrobbleRepository)(nil)

// --- PluginRepository ---

type pluginRepository struct{}

func NewPluginRepository(_ context.Context) model.PluginRepository {
	return &pluginRepository{}
}

func (r *pluginRepository) ClearErrors() error {
	return nil
}

func (r *pluginRepository) CountAll(_ ...model.QueryOptions) (int64, error) {
	return 0, nil
}

func (r *pluginRepository) Delete(_ string) error {
	return nil
}

func (r *pluginRepository) Get(_ string) (*model.Plugin, error) {
	return nil, model.ErrNotFound
}

func (r *pluginRepository) GetAll(_ ...model.QueryOptions) (model.Plugins, error) {
	return model.Plugins{}, nil
}

func (r *pluginRepository) Put(_ *model.Plugin) error {
	return nil
}

func (r *pluginRepository) Count(_ ...rest.QueryOptions) (int64, error) {
	return 0, nil
}

func (r *pluginRepository) Read(_ string) (any, error) {
	return nil, model.ErrNotFound
}

func (r *pluginRepository) ReadAll(_ ...rest.QueryOptions) (any, error) {
	return model.Plugins{}, nil
}

func (r *pluginRepository) EntityName() string {
	return "plugin"
}

func (r *pluginRepository) NewInstance() any {
	return &model.Plugin{}
}

func (r *pluginRepository) Save(_ any) (string, error) {
	return "", nil
}

func (r *pluginRepository) Update(_ string, _ any, _ ...string) error {
	return nil
}

var _ model.PluginRepository = (*pluginRepository)(nil)
var _ rest.Repository = (*pluginRepository)(nil)
var _ rest.Persistable = (*pluginRepository)(nil)

// --- FolderRepository ---

type folderRepository struct {
	ctx     context.Context
	queries *pgdb.Queries
	pool    *pgxpool.Pool
}

func NewFolderRepository(ctx context.Context, q *pgdb.Queries, p *pgxpool.Pool) model.FolderRepository {
	return &folderRepository{ctx: ctx, queries: q, pool: p}
}

func (r *folderRepository) Get(_ string) (*model.Folder, error) {
	return nil, model.ErrNotFound
}

func (r *folderRepository) GetByPath(_ model.Library, _ string) (*model.Folder, error) {
	return nil, model.ErrNotFound
}

func (r *folderRepository) GetAll(_ ...model.QueryOptions) ([]model.Folder, error) {
	return []model.Folder{}, nil
}

func (r *folderRepository) CountAll(_ ...model.QueryOptions) (int64, error) {
	return 0, nil
}

func (r *folderRepository) GetFolderUpdateInfo(_ model.Library, _ ...string) (map[string]model.FolderUpdateInfo, error) {
	return map[string]model.FolderUpdateInfo{}, nil
}

func (r *folderRepository) HasAudioOutsideFolders(_ model.Folder, _ []string) (bool, error) {
	return false, nil
}

func (r *folderRepository) Put(_ *model.Folder) error {
	return nil
}

func (r *folderRepository) MarkMissing(_ bool, _ ...string) error {
	return nil
}

func (r *folderRepository) GetTouchedWithPlaylists() (model.FolderCursor, error) {
	return func(yield func(model.Folder, error) bool) {}, nil
}

func (r *folderRepository) GetAllWithPlaylists() (model.FolderCursor, error) {
	return func(yield func(model.Folder, error) bool) {}, nil
}

var _ model.FolderRepository = (*folderRepository)(nil)

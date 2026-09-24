package postgres

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	dbschema "github.com/navidrome/navidrome/db/postgres"
	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/persistence/postgres/pgdb"
)

type PostgresStore struct {
	pool    *pgxpool.Pool
	queries *pgdb.Queries
}

// New creates a new PostgresStore connected to Neon/Postgres using pgxpool.
func New(ctx context.Context, connString string) (*PostgresStore, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parsing postgres connection string: %w", err)
	}

	// Serverless-friendly connection pool tuning
	cfg.MaxConns = 10
	cfg.MinConns = 0
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}

	// Automatically ensure all database tables and indexes exist on connect
	if _, err := pool.Exec(ctx, dbschema.SchemaSQL); err != nil {
		return nil, fmt.Errorf("initializing postgres database schema: %w", err)
	}

	// Clean up any incomplete/corrupted user created with empty password
	_, _ = pool.Exec(ctx, `DELETE FROM "user" WHERE password = '';`)

	return &PostgresStore{
		pool:    pool,
		queries: pgdb.New(pool),
	}, nil
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *PostgresStore) Queries() *pgdb.Queries {
	return s.queries
}

func (s *PostgresStore) User(ctx context.Context) model.UserRepository {
	return NewUserRepository(ctx, s.queries, s.pool)
}

func (s *PostgresStore) Library(ctx context.Context) model.LibraryRepository {
	return NewLibraryRepository(ctx, s.queries, s.pool)
}

func (s *PostgresStore) Property(ctx context.Context) model.PropertyRepository {
	return NewPropertyRepository(ctx, s.queries)
}

func (s *PostgresStore) Folder(ctx context.Context) model.FolderRepository {
	return NewFolderRepository(ctx, s.queries, s.pool)
}

func (s *PostgresStore) Album(ctx context.Context) model.AlbumRepository {
	return NewAlbumRepository(ctx, s.queries, s.pool)
}

func (s *PostgresStore) Artist(ctx context.Context) model.ArtistRepository {
	return NewArtistRepository(ctx, s.queries, s.pool)
}

func (s *PostgresStore) MediaFile(ctx context.Context) model.MediaFileRepository {
	return NewMediaFileRepository(ctx, s.queries, s.pool)
}

func (s *PostgresStore) Genre(ctx context.Context) model.GenreRepository {
	return NewGenreRepository(ctx, s.pool)
}

func (s *PostgresStore) Tag(ctx context.Context) model.TagRepository {
	return NewTagRepository(ctx, s.pool)
}

func (s *PostgresStore) Playlist(ctx context.Context) model.PlaylistRepository {
	return NewPlaylistRepository(ctx, s.queries, s.pool)
}

func (s *PostgresStore) PlayQueue(ctx context.Context) model.PlayQueueRepository {
	return NewPlayQueueRepository(ctx, s.pool)
}

func (s *PostgresStore) Transcoding(ctx context.Context) model.TranscodingRepository {
	return NewTranscodingRepository(ctx)
}

func (s *PostgresStore) Player(ctx context.Context) model.PlayerRepository {
	return NewPlayerRepository(ctx, s.pool)
}

func (s *PostgresStore) Radio(ctx context.Context) model.RadioRepository {
	return NewRadioRepository(ctx, s.pool)
}

func (s *PostgresStore) Share(ctx context.Context) model.ShareRepository {
	return NewShareRepository(ctx, s.pool)
}

func (s *PostgresStore) UserProps(ctx context.Context) model.UserPropsRepository {
	return NewUserPropsRepository(ctx, s.pool)
}

func (s *PostgresStore) ScrobbleBuffer(ctx context.Context) model.ScrobbleBufferRepository {
	return NewScrobbleBufferRepository(ctx)
}

func (s *PostgresStore) Scrobble(ctx context.Context) model.ScrobbleRepository {
	return NewScrobbleRepository(ctx, s.pool)
}

func (s *PostgresStore) Plugin(ctx context.Context) model.PluginRepository {
	return NewPluginRepository(ctx)
}

func (s *PostgresStore) Artwork(ctx context.Context) model.ArtworkRepository {
	return NewArtworkRepository(ctx, s.pool)
}

func (s *PostgresStore) ArtworkQueue(ctx context.Context) model.ArtworkQueueRepository {
	return NewArtworkQueueRepository(ctx, s.pool)
}

func (s *PostgresStore) Resource(ctx context.Context, m any) model.ResourceRepository {
	switch m.(type) {
	case model.User:
		return s.User(ctx).(model.ResourceRepository)
	case model.Transcoding:
		return s.Transcoding(ctx).(model.ResourceRepository)
	case model.Player:
		return s.Player(ctx).(model.ResourceRepository)
	case model.Artist:
		return s.Artist(ctx).(model.ResourceRepository)
	case model.Album:
		return s.Album(ctx).(model.ResourceRepository)
	case model.MediaFile:
		return s.MediaFile(ctx).(model.ResourceRepository)
	case model.Genre:
		return s.Genre(ctx).(model.ResourceRepository)
	case model.Playlist:
		return s.Playlist(ctx).(model.ResourceRepository)
	case model.Radio:
		return s.Radio(ctx).(model.ResourceRepository)
	case model.Share:
		return s.Share(ctx).(model.ResourceRepository)
	case model.Tag:
		return s.Tag(ctx).(model.ResourceRepository)
	case model.Plugin:
		return s.Plugin(ctx).(model.ResourceRepository)
	case model.Scrobble:
		return s.Scrobble(ctx).(model.ResourceRepository)
	}
	log.Error("Resource not implemented for postgres", "model", reflect.TypeOf(m).Name())
	return nil
}

// Transaction support
func (s *PostgresStore) WithTx(block func(tx model.DataStore) error, _ ...string) error {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txStore := &PostgresStore{
		pool:    s.pool,
		queries: s.queries.WithTx(tx),
	}

	if err := block(txStore); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *PostgresStore) WithTxImmediate(block func(tx model.DataStore) error, scope ...string) error {
	return s.WithTx(block, scope...)
}

func (s *PostgresStore) GC(ctx context.Context, libraryIDs ...int) error {
	// Garbage collection for missing records
	return nil
}

var _ model.DataStore = (*PostgresStore)(nil)

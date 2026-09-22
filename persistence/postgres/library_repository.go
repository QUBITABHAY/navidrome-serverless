package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/deluan/rest"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/persistence/postgres/pgdb"
)

type libraryRepository struct {
	ctx     context.Context
	queries *pgdb.Queries
	pool    *pgxpool.Pool
}

func NewLibraryRepository(ctx context.Context, q *pgdb.Queries, p *pgxpool.Pool) model.LibraryRepository {
	return &libraryRepository{
		ctx:     ctx,
		queries: q,
		pool:    p,
	}
}

func toModelLibrary(l pgdb.Library) *model.Library {
	return &model.Library{
		ID:                 int(l.ID),
		Name:               l.Name,
		Path:               l.Path,
		RemotePath:         l.RemotePath,
		LastScanAt:         l.LastScanAt.Time,
		LastScanStartedAt:  l.LastScanStartedAt.Time,
		FullScanInProgress: l.FullScanInProgress,
		TotalSongs:         int(l.TotalSongs),
		TotalAlbums:        int(l.TotalAlbums),
		TotalArtists:       int(l.TotalArtists),
		TotalFolders:       int(l.TotalFolders),
		TotalFiles:         int(l.TotalFiles),
		TotalMissingFiles:  int(l.TotalMissingFiles),
		TotalSize:          l.TotalSize,
		TotalDuration:      float64(l.TotalDuration),
		DefaultNewUsers:    l.DefaultNewUsers,
		CreatedAt:          l.CreatedAt.Time,
		UpdatedAt:          l.UpdatedAt.Time,
	}
}

func (r *libraryRepository) Get(id int) (*model.Library, error) {
	l, err := r.queries.GetLibraryByID(r.ctx, int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return toModelLibrary(l), nil
}

func (r *libraryRepository) GetPath(id int) (string, error) {
	return r.queries.GetLibraryPath(r.ctx, int32(id))
}

func (r *libraryRepository) GetAll(_ ...model.QueryOptions) (model.Libraries, error) {
	dbLibs, err := r.queries.ListLibraries(r.ctx)
	if err != nil {
		return nil, err
	}
	libs := make(model.Libraries, len(dbLibs))
	for i, l := range dbLibs {
		libs[i] = *toModelLibrary(l)
	}
	return libs, nil
}

func (r *libraryRepository) CountAll(_ ...model.QueryOptions) (int64, error) {
	return r.queries.CountLibraries(r.ctx)
}

func (r *libraryRepository) Put(l *model.Library, _ ...string) error {
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now()
	}
	l.UpdatedAt = time.Now()

	params := pgdb.UpsertLibraryParams{
		ID:              int32(l.ID),
		Name:            l.Name,
		Path:            l.Path,
		RemotePath:      l.RemotePath,
		DefaultNewUsers: l.DefaultNewUsers,
		CreatedAt:       pgtype.Timestamptz{Time: l.CreatedAt, Valid: true},
		UpdatedAt:       pgtype.Timestamptz{Time: l.UpdatedAt, Valid: true},
	}

	saved, err := r.queries.UpsertLibrary(r.ctx, params)
	if err != nil {
		return err
	}
	*l = *toModelLibrary(saved)
	return nil
}

func (r *libraryRepository) Delete(id int) error {
	return r.queries.DeleteLibrary(r.ctx, int32(id))
}

func (r *libraryRepository) StoreMusicFolder() error {
	return nil
}

func (r *libraryRepository) AddArtist(id int, artistID string) error {
	return r.queries.AddLibraryArtist(r.ctx, pgdb.AddLibraryArtistParams{
		LibraryID: int32(id),
		ArtistID:  artistID,
		Stats:     "{}",
	})
}

func (r *libraryRepository) GetUsersWithLibraryAccess(libraryID int) (model.Users, error) {
	dbUsers, err := r.queries.GetUsersWithLibraryAccess(r.ctx, int32(libraryID))
	if err != nil {
		return nil, err
	}
	users := make(model.Users, len(dbUsers))
	for i, u := range dbUsers {
		users[i] = *toModelUser(u)
	}
	return users, nil
}

func (r *libraryRepository) ScanBegin(id int, fullScan bool) error {
	return r.queries.ScanBegin(r.ctx, pgdb.ScanBeginParams{
		ID:                 int32(id),
		FullScanInProgress: fullScan,
	})
}

func (r *libraryRepository) ScanEnd(id int) error {
	return r.queries.ScanEnd(r.ctx, int32(id))
}

func (r *libraryRepository) ScanInProgress() (bool, error) {
	return r.queries.ScanInProgress(r.ctx)
}

func (r *libraryRepository) RefreshStats(id int) error {
	return r.queries.RefreshLibraryStats(r.ctx, int32(id))
}

// REST methods
func (r *libraryRepository) Count(_ ...rest.QueryOptions) (int64, error) {
	return r.CountAll()
}

func (r *libraryRepository) Read(id string) (any, error) {
	var intID int
	if _, err := parseID(id, &intID); err != nil {
		return nil, err
	}
	return r.Get(intID)
}

func (r *libraryRepository) ReadAll(_ ...rest.QueryOptions) (any, error) {
	return r.GetAll()
}

func (r *libraryRepository) Save(entity any) (string, error) {
	l, ok := entity.(*model.Library)
	if !ok {
		return "", errors.New("invalid library entity")
	}
	if err := r.Put(l); err != nil {
		return "", err
	}
	return idToString(l.ID), nil
}

func (r *libraryRepository) Update(id string, entity any, _ ...string) error {
	l, ok := entity.(*model.Library)
	if !ok {
		return errors.New("invalid library entity")
	}
	var intID int
	if _, err := parseID(id, &intID); err != nil {
		return err
	}
	l.ID = intID
	return r.Put(l)
}

func (r *libraryRepository) EntityName() string {
	return "library"
}

func (r *libraryRepository) NewInstance() any {
	return &model.Library{}
}

var _ model.LibraryRepository = (*libraryRepository)(nil)

package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/deluan/rest"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/navidrome/navidrome/conf"
	"github.com/navidrome/navidrome/consts"
	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/model/id"
	"github.com/navidrome/navidrome/persistence/postgres/pgdb"
	"github.com/navidrome/navidrome/utils"
)

type userRepository struct {
	ctx     context.Context
	queries *pgdb.Queries
	pool    *pgxpool.Pool
}

func NewUserRepository(ctx context.Context, q *pgdb.Queries, p *pgxpool.Pool) model.UserRepository {
	return &userRepository{
		ctx:     ctx,
		queries: q,
		pool:    p,
	}
}

func toModelUser(u pgdb.User) *model.User {
	user := &model.User{
		ID:             u.ID,
		UserName:       u.UserName,
		Name:           u.Name,
		Email:          u.Email,
		Password:       u.Password,
		IsAdmin:        u.IsAdmin,
		TokenEpoch:     int(u.TokenEpoch),
		ScrobbleFilter: u.ScrobbleFilter,
		CreatedAt:      u.CreatedAt.Time,
		UpdatedAt:      u.UpdatedAt.Time,
	}
	if u.LastLoginAt.Valid {
		t := u.LastLoginAt.Time
		user.LastLoginAt = &t
	}
	if u.LastAccessAt.Valid {
		t := u.LastAccessAt.Time
		user.LastAccessAt = &t
	}
	return user
}

func (r *userRepository) CountAll(_ ...model.QueryOptions) (int64, error) {
	return r.queries.CountUsers(r.ctx)
}

func (r *userRepository) Delete(id string) error {
	return r.queries.DeleteUser(r.ctx, id)
}

func (r *userRepository) Get(id string) (*model.User, error) {
	u, err := r.queries.GetUserByID(r.ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	res := toModelUser(u)
	libs, err := r.GetUserLibraries(id)
	if err == nil {
		res.Libraries = libs
	}
	return res, nil
}

func (r *userRepository) GetAll(_ ...model.QueryOptions) (model.Users, error) {
	dbUsers, err := r.queries.ListUsers(r.ctx)
	if err != nil {
		return nil, err
	}
	users := make(model.Users, len(dbUsers))
	for i, u := range dbUsers {
		users[i] = *toModelUser(u)
	}
	return users, nil
}

func keyTo32Bytes(input string) []byte {
	data := sha256.Sum256([]byte(input))
	return data[0:]
}

func getEncryptionKey() []byte {
	if conf.Server.PasswordEncryptionKey != "" {
		return keyTo32Bytes(conf.Server.PasswordEncryptionKey)
	}
	return keyTo32Bytes(consts.DefaultEncryptionKey)
}

func (r *userRepository) encryptPassword(u *model.User) error {
	encKey := getEncryptionKey()
	encPassword, err := utils.Encrypt(r.ctx, encKey, u.NewPassword)
	if err != nil {
		log.Error(r.ctx, "Error encrypting user's password", "user", u.UserName, err)
		return err
	}
	u.NewPassword = encPassword
	return nil
}

func (r *userRepository) decryptPassword(u *model.User) error {
	if u.Password == "" {
		return nil
	}
	encKey := getEncryptionKey()
	plaintext, err := utils.Decrypt(r.ctx, encKey, u.Password)
	if err != nil {
		log.Error(r.ctx, "Error decrypting user's password", "user", u.UserName, err)
		return err
	}
	u.Password = plaintext
	return nil
}

func (r *userRepository) Put(u *model.User) error {
	if u.ID == "" {
		u.ID = id.NewRandom()
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	u.UpdatedAt = time.Now()

	passwordToSave := u.Password
	if u.NewPassword != "" {
		if err := r.encryptPassword(u); err != nil {
			return err
		}
		passwordToSave = u.NewPassword
		u.TokenEpoch++
	}

	params := pgdb.UpsertUserParams{
		ID:             u.ID,
		UserName:       u.UserName,
		Name:           u.Name,
		Email:          u.Email,
		Password:       passwordToSave,
		IsAdmin:        u.IsAdmin,
		TokenEpoch:     int32(u.TokenEpoch),
		ScrobbleFilter: u.ScrobbleFilter,
		CreatedAt:      pgtype.Timestamptz{Time: u.CreatedAt, Valid: true},
		UpdatedAt:      pgtype.Timestamptz{Time: u.UpdatedAt, Valid: true},
	}

	saved, err := r.queries.UpsertUser(r.ctx, params)
	if err != nil {
		return fmt.Errorf("upserting user: %w", err)
	}
	*u = *toModelUser(saved)

	if u.IsAdmin {
		_, _ = r.pool.Exec(r.ctx,
			`INSERT INTO user_library (user_id, library_id) SELECT $1, id FROM library ON CONFLICT DO NOTHING`,
			u.ID,
		)
	}
	return nil
}

func (r *userRepository) UpdateLastLoginAt(id string) error {
	return r.queries.UpdateLastLoginAt(r.ctx, pgdb.UpdateLastLoginAtParams{
		ID:          id,
		LastLoginAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
}

func (r *userRepository) UpdateLastAccessAt(id string) error {
	return r.queries.UpdateLastAccessAt(r.ctx, pgdb.UpdateLastAccessAtParams{
		ID:           id,
		LastAccessAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
}

func (r *userRepository) FindFirstAdmin() (*model.User, error) {
	u, err := r.queries.FindFirstAdmin(r.ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return toModelUser(u), nil
}

func (r *userRepository) FindByUsername(username string) (*model.User, error) {
	u, err := r.queries.GetUserByUsername(r.ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	res := toModelUser(u)
	libs, err := r.GetUserLibraries(res.ID)
	if err == nil {
		res.Libraries = libs
	}
	return res, nil
}

func (r *userRepository) FindByUsernameWithPassword(username string) (*model.User, error) {
	usr, err := r.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if usr != nil {
		_ = r.decryptPassword(usr)
	}
	return usr, nil
}

func (r *userRepository) GetUserLibraries(userID string) (model.Libraries, error) {
	dbLibs, err := r.queries.GetUserLibraries(r.ctx, userID)
	if err != nil {
		return nil, err
	}
	libs := make(model.Libraries, len(dbLibs))
	for i, l := range dbLibs {
		libs[i] = *toModelLibrary(l)
	}
	return libs, nil
}

func (r *userRepository) SetUserLibraries(userID string, libraryIDs []int) error {
	if err := r.queries.ClearUserLibraries(r.ctx, userID); err != nil {
		return err
	}
	for _, libID := range libraryIDs {
		if err := r.queries.AddUserLibrary(r.ctx, pgdb.AddUserLibraryParams{
			UserID:    userID,
			LibraryID: int32(libID),
		}); err != nil {
			return err
		}
	}
	return nil
}

// REST Repository integration
func (r *userRepository) Count(_ ...rest.QueryOptions) (int64, error) {
	return r.CountAll()
}

func (r *userRepository) Read(id string) (any, error) {
	return r.Get(id)
}

func (r *userRepository) ReadAll(_ ...rest.QueryOptions) (any, error) {
	return r.GetAll()
}

func (r *userRepository) Save(entity any) (string, error) {
	u, ok := entity.(*model.User)
	if !ok {
		return "", errors.New("invalid user entity")
	}
	if err := r.Put(u); err != nil {
		return "", err
	}
	return u.ID, nil
}

func (r *userRepository) Update(id string, entity any, _ ...string) error {
	u, ok := entity.(*model.User)
	if !ok {
		return errors.New("invalid user entity")
	}
	u.ID = id
	return r.Put(u)
}

func (r *userRepository) EntityName() string {
	return "user"
}

func (r *userRepository) NewInstance() any {
	return &model.User{}
}

var _ model.UserRepository = (*userRepository)(nil)

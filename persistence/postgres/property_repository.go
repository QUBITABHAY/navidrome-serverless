package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/persistence/postgres/pgdb"
)

type propertyRepository struct {
	ctx     context.Context
	queries *pgdb.Queries
}

func NewPropertyRepository(ctx context.Context, q *pgdb.Queries) model.PropertyRepository {
	return &propertyRepository{
		ctx:     ctx,
		queries: q,
	}
}

func (r *propertyRepository) Put(id string, value string) error {
	if r.queries == nil {
		return nil
	}
	return r.queries.PutProperty(r.ctx, pgdb.PutPropertyParams{
		ID:    id,
		Value: value,
	})
}

func (r *propertyRepository) Get(id string) (string, error) {
	if r.queries == nil {
		return "", model.ErrNotFound
	}
	val, err := r.queries.GetProperty(r.ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", model.ErrNotFound
		}
		return "", err
	}
	return val, nil
}

func (r *propertyRepository) Delete(id string) error {
	if r.queries == nil {
		return nil
	}
	return r.queries.DeleteProperty(r.ctx, id)
}

func (r *propertyRepository) DefaultGet(id string, defaultValue string) (string, error) {
	val, err := r.Get(id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return defaultValue, nil
		}
		return "", err
	}
	return val, nil
}

var _ model.PropertyRepository = (*propertyRepository)(nil)

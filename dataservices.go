package dataservicex

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
)

type model interface {
	IDColumnName() string
	TableName() string
}

type DataServices[T model] struct {
	db      *sqlx.DB
	dialect goqu.DialectWrapper
}

func NewDataServices[T model](db *sqlx.DB, options ...func(*DataServices[T])) DataServices[T] {
	ds := DataServices[T]{
		db:      db,
		dialect: goqu.Dialect("default"),
	}
	for _, option := range options {
		option(&ds)
	}
	return ds
}

func WithDialect[T model](dialect goqu.DialectWrapper) func(*DataServices[T]) {
	return func(s *DataServices[T]) {
		s.dialect = dialect
	}
}

func (ds *DataServices[T]) GetList(ctx context.Context) ([]T, error) {
	var t T
	query, params, err := ds.dialect.
		From(t.TableName()).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("fail to build select sql script: %w", err)
	}

	var ms []T
	return ms, ds.db.SelectContext(ctx, &ms, query, params...)
}

func (ds *DataServices[T]) GetByID(ctx context.Context, id interface{}) (T, error) {
	var t T
	query, params, err := ds.dialect.
		From(t.TableName()).
		Where(goqu.Ex{t.IDColumnName(): id}).
		ToSQL()
	if err != nil {
		return t, fmt.Errorf("fail to build get sql script: %w", err)
	}

	return t, ds.db.GetContext(ctx, &t, query, params...)
}

func (ds *DataServices[T]) Create(ctx context.Context, m T) (T, error) {
	q := ds.dialect.
		Insert(m.TableName()).
		Rows(m)

	if ds.db.DriverName() == "postgres" {
		q = q.Returning("id")
	}

	query, params, err := q.ToSQL()
	if err != nil {
		return m, fmt.Errorf("fail to build insert sql script: %w", err)
	}

	var insertedID int64

	if ds.db.DriverName() == "postgres" {
		// docs: https://pkg.go.dev/github.com/lib/pq#hdr-Queries
		err = ds.db.QueryRowxContext(ctx, query, params...).Scan(&insertedID)
		if err != nil {
			return m, fmt.Errorf("fail to execute insert sql script: %w", err)
		}
	} else {
		result, err := ds.db.ExecContext(ctx, query, params...)
		if err != nil {
			return m, fmt.Errorf("fail to execute insert sql script: %w", err)
		}

		insertedID, err = result.LastInsertId()
		if err != nil {
			return m, fmt.Errorf("query affected id of insert sql script: %w", err)
		}
	}

	return ds.GetByID(ctx, insertedID)
}

func (ds *DataServices[T]) Update(ctx context.Context, id interface{}, updating T) (T, error) {
	query, params, err := ds.dialect.
		Update(updating.TableName()).
		Set(updating).
		Where(goqu.Ex{updating.IDColumnName(): id}).
		ToSQL()
	if err != nil {
		return updating, fmt.Errorf("fail to build update sql script: %w", err)
	}
	_, err = ds.db.ExecContext(ctx, query, params...)
	if err != nil {
		return updating, fmt.Errorf("fail to execute update sql script: %w", err)
	}
	return ds.GetByID(ctx, id)
}

func (ds *DataServices[T]) UpdateColumns(ctx context.Context, id interface{}, record goqu.Record) error {
	var m T
	query, params, err := ds.dialect.
		Update(m.TableName()).
		Set(record).
		Where(goqu.Ex{m.IDColumnName(): id}).
		ToSQL()
	if err != nil {
		return fmt.Errorf("fail to build update sql script: %w", err)
	}

	_, err = ds.db.ExecContext(ctx, query, params...)
	if err != nil {
		return fmt.Errorf("fail to execute update sql script: %w", err)
	}
	return nil
}

func (ds *DataServices[T]) Delete(ctx context.Context, id interface{}) error {
	var t T
	query, params, err := ds.dialect.
		Delete(t.TableName()).
		Where(goqu.Ex{t.IDColumnName(): id}).
		ToSQL()
	if err != nil {
		return fmt.Errorf("fail to build update sql script: %w", err)
	}

	_, err = ds.db.ExecContext(ctx, query, params...)
	return err
}

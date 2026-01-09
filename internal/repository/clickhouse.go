package repository

import (
	"context"
	"fmt"

	"github.com/uptrace/go-clickhouse/ch"
)

type IClickHouseRepository[T any] interface {
	FindAll(ctx context.Context) ([]T, error)
	FindOneById(ctx context.Context, id interface{}) (*T, error)
	FindByIds(ctx context.Context, ids []interface{}) ([]T, error)
	FindWhere(ctx context.Context, condition string, args ...interface{}) ([]T, error)
	Create(ctx context.Context, model *T) error
	BulkCreate(ctx context.Context, inputs []T) error
	Count(ctx context.Context) (int64, error)
	CountWhere(ctx context.Context, condition string, args ...interface{}) (int64, error)
}

type ClickHouseRepository[T any] struct {
	db *ch.DB
}

var _ IClickHouseRepository[any] = &ClickHouseRepository[any]{}

func (r *ClickHouseRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	var results []T
	err := r.db.NewSelect().
		Model(&results).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *ClickHouseRepository[T]) FindOneById(ctx context.Context, id interface{}) (*T, error) {
	var result T
	err := r.db.NewSelect().
		Model(&result).
		Where("id = ?", id).
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *ClickHouseRepository[T]) FindByIds(ctx context.Context, ids []interface{}) ([]T, error) {
	var results []T
	err := r.db.NewSelect().
		Model(&results).
		Where("id IN (?)", ch.In(ids)).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *ClickHouseRepository[T]) FindWhere(ctx context.Context, condition string, args ...interface{}) ([]T, error) {
	var results []T
	err := r.db.NewSelect().
		Model(&results).
		Where(condition, args...).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *ClickHouseRepository[T]) Create(ctx context.Context, model *T) error {
	_, err := r.db.NewInsert().
		Model(model).
		Exec(ctx)
	return err
}

func (r *ClickHouseRepository[T]) BulkCreate(ctx context.Context, inputs []T) error {
	if len(inputs) == 0 {
		return nil
	}

	_, err := r.db.NewInsert().
		Model(&inputs).
		Exec(ctx)
	return err
}

func (r *ClickHouseRepository[T]) Count(ctx context.Context) (int64, error) {
	var result T
	count, err := r.db.NewSelect().
		Model(&result).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}

func (r *ClickHouseRepository[T]) CountWhere(ctx context.Context, condition string, args ...interface{}) (int64, error) {
	var result T
	count, err := r.db.NewSelect().
		Model(&result).
		Where(condition, args...).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}

type QueryBuilder[T any] struct {
	query *ch.SelectQuery
}

// NewQueryBuilder creates a new query builder
func (r *ClickHouseRepository[T]) NewQueryBuilder(ctx context.Context) *QueryBuilder[T] {
	var model T
	return &QueryBuilder[T]{
		query: r.db.NewSelect().Model(&model),
	}
}

func (qb *QueryBuilder[T]) Where(condition string, args ...interface{}) *QueryBuilder[T] {
	qb.query = qb.query.Where(condition, args...)
	return qb
}

func (qb *QueryBuilder[T]) OrderBy(fields ...string) *QueryBuilder[T] {
	qb.query = qb.query.Order(fields...)
	return qb
}

func (qb *QueryBuilder[T]) Limit(n int) *QueryBuilder[T] {
	qb.query = qb.query.Limit(n)
	return qb
}

func (qb *QueryBuilder[T]) Offset(n int) *QueryBuilder[T] {
	qb.query = qb.query.Offset(n)
	return qb
}

func (qb *QueryBuilder[T]) GroupBy(fields ...string) *QueryBuilder[T] {
	qb.query = qb.query.Group(fields...)
	return qb
}

func (qb *QueryBuilder[T]) Execute(ctx context.Context) ([]T, error) {
	var results []T
	qb.query.Model(&results)
	err := qb.query.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (qb *QueryBuilder[T]) ExecuteOne(ctx context.Context) (*T, error) {
	var result T
	qb.query.Model(&result).Limit(1)
	err := qb.query.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (qb *QueryBuilder[T]) Count(ctx context.Context) (int64, error) {
	count, err := qb.query.Count(ctx)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}

func (r *ClickHouseRepository[T]) FindWithPagination(ctx context.Context, page, pageSize int) ([]T, int64, error) {
	var results []T
	var model T

	// Get total count
	total, err := r.db.NewSelect().
		Model(&model).
		Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err = r.db.NewSelect().
		Model(&results).
		Limit(pageSize).
		Offset(offset).
		Scan(ctx)
	if err != nil {
		return nil, 0, err
	}

	return results, int64(total), nil
}

// Aggregate performs aggregation query
type AggregateResult struct {
	Value interface{} `ch:"value"`
}

func (r *ClickHouseRepository[T]) Aggregate(ctx context.Context, aggregateFunc, field string) (interface{}, error) {
	var result AggregateResult
	var model T

	query := fmt.Sprintf("%s(%s) as value", aggregateFunc, field)
	err := r.db.NewSelect().
		Model(&model).
		ColumnExpr(query).
		Scan(ctx, &result)

	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

func (r *ClickHouseRepository[T]) RawQuery(ctx context.Context, query string, args ...interface{}) ([]T, error) {
	var results []T
	err := r.db.NewRaw(query, args...).Scan(ctx, &results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *ClickHouseRepository[T]) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := r.db.Exec(query, args...)
	return err
}

package repository

import (
	"context"

	"github.com/hiamthach108/simplerank/internal/dto"
	"github.com/hiamthach108/simplerank/internal/model"
	"github.com/uptrace/go-clickhouse/ch"
)

type IHistoryRepository interface {
	IClickHouseRepository[model.History]
	GetList(ctx context.Context, req dto.ListHistoriesReq) ([]model.History, int64, error)
}

type HistoryRepository struct {
	ClickHouseRepository[model.History]
}

func NewHistoryRepository(dbClient *ch.DB) IHistoryRepository {
	return &HistoryRepository{
		ClickHouseRepository: ClickHouseRepository[model.History]{
			db: dbClient,
		},
	}
}

// GetList retrieves a list of history records based on the provided request filters.
func (r *HistoryRepository) GetList(ctx context.Context, req dto.ListHistoriesReq) ([]model.History, int64, error) {
	return []model.History{}, 0, nil
}

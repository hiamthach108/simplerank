package repository

import (
	"context"

	"github.com/hiamthach108/simplerank/internal/dto"
	"github.com/hiamthach108/simplerank/internal/model"
	"gorm.io/gorm"
)

type IHistoryRepository interface {
	IRepository[model.History]
	GetList(ctx context.Context, req dto.ListHistoriesReq) ([]model.History, int64, error)
}

type HistoryRepository struct {
	Repository[model.History]
}

func NewHistoryRepository(dbClient *gorm.DB) IHistoryRepository {
	return &HistoryRepository{
		Repository: Repository[model.History]{dbClient: dbClient},
	}
}

// GetList retrieves a list of history records based on the provided request filters.
func (r *HistoryRepository) GetList(ctx context.Context, req dto.ListHistoriesReq) ([]model.History, int64, error) {
	var results []model.History
	db := r.dbClient.WithContext(ctx).Model(&model.History{}).Where("leaderboard_id = ?", req.LeaderboardID)

	if req.EntryID != nil {
		db = db.Where("entry_id = ?", *req.EntryID)
	}
	if req.FromScore != nil {
		db = db.Where("score >= ?", *req.FromScore)
	}
	if req.ToScore != nil {
		db = db.Where("score <= ?", *req.ToScore)
	}
	if req.FromDate != nil {
		db = db.Where("created_at >= ?", *req.FromDate)
	}
	if req.ToDate != nil {
		db = db.Where("created_at <= ?", *req.ToDate)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").Offset(req.Page - 1).Limit(req.PageSize).Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

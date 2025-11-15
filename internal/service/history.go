package service

import (
	"context"

	"github.com/hiamthach108/simplerank/internal/dto"
	"github.com/hiamthach108/simplerank/internal/model"
	"github.com/hiamthach108/simplerank/internal/repository"
	"github.com/hiamthach108/simplerank/pkg/cache"
	"github.com/hiamthach108/simplerank/pkg/logger"
)

type IHistorySvc interface {
	Record(ctx context.Context, req *dto.CreateHistoryReq) (*model.History, error)
	List(ctx context.Context, req *dto.ListHistoriesReq) (*dto.PaginationResp[dto.HistoryDto], error)
}

type HistorySvc struct {
	logger      logger.ILogger
	cache       cache.ICache
	historyRepo repository.IHistoryRepository
}

func NewHistorySvc(logger logger.ILogger, cache cache.ICache, historyRepo repository.IHistoryRepository) IHistorySvc {
	return &HistorySvc{
		logger:      logger,
		cache:       cache,
		historyRepo: historyRepo,
	}
}

// Record saves a score history record.
func (s *HistorySvc) Record(ctx context.Context, req *dto.CreateHistoryReq) (*model.History, error) {
	m := req.ToModel()

	err := s.historyRepo.Create(ctx, m)
	if err != nil {
		s.logger.Error("[HistorySvc] failed to record score history", "leaderboard", req.LeaderboardID, "entry", req.EntryID, "error", err)
		return nil, err
	}

	return m, nil
}

// List retrieves a list of score history records based on the provided request filters.
func (s *HistorySvc) List(ctx context.Context, req *dto.ListHistoriesReq) (*dto.PaginationResp[dto.HistoryDto], error) {
	histories, total, err := s.historyRepo.GetList(ctx, *req)
	if err != nil {
		s.logger.Error("[HistorySvc] failed to list score histories", "leaderboard", req.LeaderboardID, "error", err)
		return nil, err
	}

	var historyDtos []dto.HistoryDto
	for _, history := range histories {
		historyDtos = append(historyDtos, dto.HistoryDto{}.FromModel(&history))
	}

	return &dto.PaginationResp[dto.HistoryDto]{
		Items:    historyDtos,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

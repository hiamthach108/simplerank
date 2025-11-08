package service

import (
	"context"
	"fmt"

	"github.com/hiamthach108/simplerank/internal/dto"
	"github.com/hiamthach108/simplerank/internal/model"
	"github.com/hiamthach108/simplerank/internal/repository"
	"github.com/hiamthach108/simplerank/pkg/cache"
	"github.com/hiamthach108/simplerank/pkg/logger"
)

type ILeaderboardSvc interface {
	GetLeaderboardDetail(ctx context.Context, leaderboardID string) (*dto.LeaderboardDto, error)
	GetEntryRank(ctx context.Context, leaderboardID string, entryID string) (int, error)
	UpdateEntryScore(ctx context.Context, leaderboardID string, entryID string, score float64) error
	GetListLeaderboards(ctx context.Context) ([]dto.LeaderboardDto, error)
	CreateLeaderboard(ctx context.Context, req dto.CreateLeaderboardReq) (*dto.LeaderboardDto, error)
	UpdateLeaderboard(ctx context.Context, leaderboardID string, req dto.UpdateLeaderboardReq) error
}

type LeaderBoardSvc struct {
	logger          logger.ILogger
	cache           cache.ICache
	leaderboardRepo repository.ILeaderboardRepository
}

func NewLeaderBoardSvc(logger logger.ILogger, cache cache.ICache, leaderboardRepo repository.ILeaderboardRepository) ILeaderboardSvc {
	return &LeaderBoardSvc{
		logger:          logger,
		cache:           cache,
		leaderboardRepo: leaderboardRepo,
	}
}

// UpdateEntryScore adds or updates an entry's score in the leaderboard.
func (s *LeaderBoardSvc) UpdateEntryScore(ctx context.Context, leaderboardID string, entryID string, score float64) error {
	leaderboard := s.leaderboardRepo.FindOneById(ctx, leaderboardID)
	if leaderboard == nil {
		s.logger.Error("[LeaderboardSvc] leaderboard not found", "id", leaderboardID)
		return fmt.Errorf("leaderboard not found")
	}

	err := s.cache.AddScore(leaderboard.ID, entryID, score)
	if err != nil {
		s.logger.Error("[LeaderboardSvc] failed to update entry score", "leaderboard", leaderboardID, "entry", entryID, "error", err)
		return fmt.Errorf("failed to update entry score: %w", err)
	}

	return nil
}

// GetTopEntries retrieves the top N entries from the leaderboard.
func (s *LeaderBoardSvc) GetLeaderboardDetail(ctx context.Context, leaderboardID string) (*dto.LeaderboardDto, error) {
	leaderboard := s.leaderboardRepo.FindOneById(ctx, leaderboardID)
	if leaderboard == nil {
		s.logger.Error("[LeaderboardSvc] leaderboard not found", "id", leaderboardID)
		return nil, fmt.Errorf("leaderboard not found")
	}

	entries, err := s.cache.GetTopN(leaderboard.ID, 100)
	if err != nil {
		s.logger.Error("[LeaderboardSvc] failed to get top entries", "leaderboard", leaderboardID, "error", err)
		return nil, fmt.Errorf("failed to get top entries: %w", err)
	}

	leaderboardDto := &dto.LeaderboardDto{}
	leaderboardDto.FromModel(leaderboard)
	leaderboardDto.TopEntries = entries

	return leaderboardDto, nil
}

// GetEntryRank retrieves an entry's rank (1-based) from the leaderboard.
func (s *LeaderBoardSvc) GetEntryRank(ctx context.Context, leaderboardID string, entryID string) (int, error) {
	rank, _, err := s.cache.GetRank(leaderboardID, entryID)
	if err != nil {
		s.logger.Error("[LeaderboardSvc] failed to get rank for entry", "leaderboard", leaderboardID, "entry", entryID, "error", err)
		return 0, fmt.Errorf("failed to get entry rank: %w", err)
	}
	return int(rank), nil
}

// GetListLeaderboards retrieves all leaderboards.
func (s *LeaderBoardSvc) GetListLeaderboards(ctx context.Context) ([]dto.LeaderboardDto, error) {
	leaderboards, err := s.leaderboardRepo.FindAll(ctx)
	if err != nil {
		s.logger.Error("[LeaderboardSvc] failed to get list of leaderboards", "error", err)
		return nil, fmt.Errorf("failed to get list of leaderboards: %w", err)
	}

	var resp []dto.LeaderboardDto
	for _, lb := range leaderboards {
		var lbDto dto.LeaderboardDto
		lbDto.FromModel(&lb)
		resp = append(resp, lbDto)
	}

	return resp, nil
}

// CreateLeaderboard creates a new leaderboard.
func (s *LeaderBoardSvc) CreateLeaderboard(ctx context.Context, req dto.CreateLeaderboardReq) (*dto.LeaderboardDto, error) {
	// Implementation depends on how leaderboards are stored.
	// This could involve creating a new key in the cache or storing metadata in a database.
	s.logger.Info("[LeaderboardSvc] Creating leaderboard", "name", req.Name)

	m := model.Leaderboard{
		Name:        req.Name,
		Description: req.Description,
		ExpiredAt:   req.ExpiredAt,
	}

	leaderboard, err := s.leaderboardRepo.Create(ctx, &m)
	if err != nil {
		s.logger.Error("[LeaderboardSvc] failed to create leaderboard", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to create leaderboard: %w", err)
	}
	var resp dto.LeaderboardDto
	resp.FromModel(leaderboard)

	return &resp, nil
}

// UpdateLeaderboard updates an existing leaderboard's details.
func (s *LeaderBoardSvc) UpdateLeaderboard(ctx context.Context, leaderboardID string, req dto.UpdateLeaderboardReq) error {
	existing := s.leaderboardRepo.FindOneById(ctx, leaderboardID)
	if existing == nil {
		s.logger.Error("[LeaderboardSvc] leaderboard not found", "id", leaderboardID)
		return fmt.Errorf("leaderboard not found")
	}

	updatedModel, fields := req.ToModel()
	if len(fields) == 0 {
		s.logger.Info("[LeaderboardSvc] no fields to update for leaderboard", "id", leaderboardID)
		return nil // Nothing to update
	}

	err := s.leaderboardRepo.Update(ctx, leaderboardID, *updatedModel, fields...)
	if err != nil {
		s.logger.Error("[LeaderboardSvc] failed to update leaderboard", "id", leaderboardID, "error", err)
		return fmt.Errorf("failed to update leaderboard: %w", err)
	}

	return nil
}

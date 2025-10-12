package service

import (
	"context"
	"fmt"

	"github.com/hiamthach108/simplerank/pkg/cache"
	"github.com/hiamthach108/simplerank/pkg/logger"
)

type ILeaderboardSvc interface {
	GetTopEntries(ctx context.Context, leaderboardID string, limit int) ([]string, error)
	GetEntryRank(ctx context.Context, leaderboardID string, entryID string) (int, error)
	UpdateEntryScore(ctx context.Context, leaderboardID string, entryID string, score float64) error
}

type LeaderBoardSvc struct {
	logger logger.ILogger
	cache  cache.ICache
}

func NewLeaderBoardSvc(logger logger.ILogger, cache cache.ICache) ILeaderboardSvc {
	return &LeaderBoardSvc{
		logger: logger,
		cache:  cache,
	}
}

// UpdateEntryScore adds or updates an entry's score in the leaderboard.
func (s *LeaderBoardSvc) UpdateEntryScore(ctx context.Context, leaderboardID string, entryID string, score float64) error {
	err := s.cache.AddScore(leaderboardID, entryID, score)
	if err != nil {
		s.logger.Error("[LeaderboardSvc] failed to update entry score", "leaderboard", leaderboardID, "entry", entryID, "error", err)
		return fmt.Errorf("failed to update entry score: %w", err)
	}
	return nil
}

// GetTopEntries retrieves the top N entries from the leaderboard.
func (s *LeaderBoardSvc) GetTopEntries(ctx context.Context, leaderboardID string, limit int) ([]string, error) {
	entries, err := s.cache.GetTopN(leaderboardID, int64(limit))
	if err != nil {
		s.logger.Error("[LeaderboardSvc] failed to get top entries", "leaderboard", leaderboardID, "error", err)
		return nil, fmt.Errorf("failed to get top entries: %w", err)
	}

	entryIDs := make([]string, 0, len(entries))
	for _, entry := range entries {
		// Defensive: ensure the member is a string
		if entryID, ok := entry.Member.(string); ok {
			entryIDs = append(entryIDs, entryID)
		} else {
			s.logger.Warn("[LeaderboardSvc] unexpected member type", "leaderboard", leaderboardID, "member", entry.Member)
		}
	}
	return entryIDs, nil
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

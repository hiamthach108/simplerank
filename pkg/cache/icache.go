package cache

import (
	"time"
)

type LeaderboardEntry struct {
	Member any     `json:"member"`
	Score  float64 `json:"score"`
}

type ICache interface {
	Set(key string, value any, expireTime *time.Duration) error
	Get(key string) (any, error)
	Delete(key string) error
	Clear() error
	ClearWithPrefix(prefix string) error
	AddScore(boardKey, member string, score float64) error
	GetTopN(boardKey string, n int64) ([]LeaderboardEntry, error)
	GetRank(boardKey, member string) (rank int64, score float64, err error)
	RemoveMember(boardKey, member string) error
	GetAroundMember(boardKey, member string, radius int64) ([]LeaderboardEntry, error)
}

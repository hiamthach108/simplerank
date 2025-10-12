package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/hiamthach108/simplerank/config"
	"github.com/hiamthach108/simplerank/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type appCache struct {
	serviceName string
	logger      logger.ILogger
	redisClient *redis.Client
}

func NewAppCache(config *config.AppConfig, logger logger.ILogger) (ICache, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Cache.RedisHost + ":" + config.Cache.RedisPort,
		Password: config.Cache.RedisPassword,
		DB:       config.Cache.RedisDB,
	})

	return &appCache{
		serviceName: config.App.Name,
		logger:      logger,
		redisClient: redisClient,
	}, nil
}

// =============================
// 🔹 Basic Cache Operations
// =============================

func (c *appCache) Set(key string, value any, expireTime *time.Duration) error {
	rKey := c.prefixedKey(key)
	return c.redisClient.Set(context.Background(), rKey, value, *expireTime).Err()
}

func (c *appCache) Get(key string) (any, error) {
	rKey := c.prefixedKey(key)
	val, err := c.redisClient.Get(context.Background(), rKey).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (c *appCache) Delete(key string) error {
	rKey := c.prefixedKey(key)
	return c.redisClient.Del(context.Background(), rKey).Err()
}

func (c *appCache) Clear() error {
	return c.redisClient.FlushAll(context.Background()).Err()
}

func (c *appCache) ClearWithPrefix(prefix string) error {
	ctx := context.Background()
	pattern := c.prefixedKey(fmt.Sprintf("%s*", prefix))
	keys, err := c.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.redisClient.Del(ctx, keys...).Err()
	}
	return nil
}

// =============================
// 🔹 Leaderboard (Sorted Set)
// =============================

// AddScore adds or updates a member’s score in a leaderboard.
func (c *appCache) AddScore(boardKey, member string, score float64) error {
	rKey := c.prefixedKey(boardKey)
	return c.redisClient.ZAdd(context.Background(), rKey, redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

// GetTopN retrieves top N members with their scores in descending order.
func (c *appCache) GetTopN(boardKey string, n int64) ([]LeaderboardEntry, error) {
	rKey := c.prefixedKey(boardKey)
	zResult, err := c.redisClient.ZRevRangeWithScores(context.Background(), rKey, 0, n-1).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]LeaderboardEntry, len(zResult))
	for i, z := range zResult {
		entries[i] = LeaderboardEntry{
			Member: z.Member,
			Score:  z.Score,
		}
	}

	return entries, nil
}

// GetRank retrieves the rank (1-based) and score of a specific member.
func (c *appCache) GetRank(boardKey, member string) (rank int64, score float64, err error) {
	rKey := c.prefixedKey(boardKey)
	rank, err = c.redisClient.ZRevRank(context.Background(), rKey, member).Result()
	if err != nil {
		return 0, 0, err
	}

	score, err = c.redisClient.ZScore(context.Background(), rKey, member).Result()
	if err != nil {
		return 0, 0, err
	}

	return rank + 1, score, nil // rank is 0-based in Redis
}

// RemoveMember removes a player from the leaderboard.
func (c *appCache) RemoveMember(boardKey, member string) error {
	rKey := c.prefixedKey(boardKey)
	return c.redisClient.ZRem(context.Background(), rKey, member).Err()
}

// GetAroundMember gets a window of players around a given member (for user’s local rank view)
func (c *appCache) GetAroundMember(boardKey, member string, radius int64) ([]LeaderboardEntry, error) {
	rKey := c.prefixedKey(boardKey)
	rank, err := c.redisClient.ZRevRank(context.Background(), rKey, member).Result()
	if err != nil {
		return nil, err
	}

	start := rank - radius
	if start < 0 {
		start = 0
	}
	end := rank + radius

	zResult, err := c.redisClient.ZRevRangeWithScores(context.Background(), rKey, start, end).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]LeaderboardEntry, len(zResult))
	for i, z := range zResult {
		entries[i] = LeaderboardEntry{
			Member: z.Member,
			Score:  z.Score,
		}
	}

	return entries, nil
}

func (c *appCache) prefixedKey(key string) string {
	return fmt.Sprintf("%s:%s", c.serviceName, key)
}

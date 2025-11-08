package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hiamthach108/simplerank/config"
	"github.com/hiamthach108/simplerank/pkg/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockLogger is a simple mock implementation of logger.ILogger
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...any) {}
func (m *MockLogger) Info(msg string, fields ...any)  {}
func (m *MockLogger) Warn(msg string, fields ...any)  {}
func (m *MockLogger) Error(msg string, fields ...any) {}
func (m *MockLogger) Fatal(msg string, fields ...any) {}
func (m *MockLogger) With(fields ...any) logger.ILogger {
	return m
}
func (m *MockLogger) GetZapLogger() *zap.Logger {
	return nil
}

// Test helper to create a test cache instance
func createTestCache() *appCache {
	mockLogger := &MockLogger{}

	// Create a test cache with mock dependencies
	cache := &appCache{
		serviceName: "test-service",
		logger:      mockLogger,
		redisClient: nil, // We'll set this in individual tests
	}

	return cache
}

func TestNewAppCache(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis connection test")
	}

	tests := []struct {
		name    string
		config  *config.AppConfig
		logger  logger.ILogger
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.AppConfig{
				App: struct {
					Name    string `env:"APP_NAME"`
					Version string `env:"APP_VERSION"`
				}{
					Name: "test-service",
				},
				Cache: struct {
					DefaultExpireTimeSec int    `env:"CACHE_DEFAULT_EXPIRE_TIME_SEC"`
					CleanupIntervalHour  int    `env:"CACHE_CLEANUP_INTERVAL_HOUR"`
					RedisHost            string `env:"REDIS_HOST"`
					RedisPort            string `env:"REDIS_PORT"`
					RedisPassword        string `env:"REDIS_PASSWORD"`
					RedisDB              int    `env:"REDIS_DB"`
				}{
					RedisHost:     "localhost",
					RedisPort:     "6379",
					RedisPassword: "",
					RedisDB:       1, // Use test DB
				},
			},
			logger:  &MockLogger{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewAppCache(tt.config, tt.logger)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cache)
			} else {
				// Redis might not be running in CI/CD
				if err != nil {
					t.Skip("Redis not available, skipping test")
				}
				assert.NoError(t, err)
				assert.NotNil(t, cache)
				assert.IsType(t, &appCache{}, cache)
			}
		})
	}
}

func TestAppCache_prefixedKey(t *testing.T) {
	cache := createTestCache()

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "simple key",
			key:      "test-key",
			expected: "test-service:test-key",
		},
		{
			name:     "empty key",
			key:      "",
			expected: "test-service:",
		},
		{
			name:     "key with special characters",
			key:      "test:key:with:colons",
			expected: "test-service:test:key:with:colons",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cache.prefixedKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test with real Redis connection (integration test)
func TestAppCache_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create a real Redis client for integration testing
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1, // Use a different DB for testing
	})

	// Test connection
	ctx := context.Background()
	err := redisClient.Ping(ctx).Err()
	if err != nil {
		t.Skip("Redis not available, skipping integration test")
	}

	// Clean up test database
	defer redisClient.FlushDB(ctx)

	// Create cache with real Redis client
	cache := &appCache{
		serviceName: "test-service",
		logger:      &MockLogger{},
		redisClient: redisClient,
	}

	t.Run("Set and Get string", func(t *testing.T) {
		key := "test-key"
		value := "test-value"
		expireTime := 5 * time.Minute

		// Test Set
		err := cache.Set(key, value, &expireTime)
		assert.NoError(t, err)

		// Test Get
		var result string
		err = cache.Get(key, &result)
		assert.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("Set and Get struct", func(t *testing.T) {
		type TestStruct struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}

		key := "test-struct-key"
		value := TestStruct{Name: "test", Count: 42}
		expireTime := 5 * time.Minute

		// Test Set
		err := cache.Set(key, value, &expireTime)
		assert.NoError(t, err)

		// Test Get
		var result TestStruct
		err = cache.Get(key, &result)
		assert.NoError(t, err)
		assert.Equal(t, value.Name, result.Name)
		assert.Equal(t, value.Count, result.Count)
	})

	t.Run("Delete", func(t *testing.T) {
		key := "test-delete-key"
		value := "test-value"
		expireTime := 5 * time.Minute

		// Set a value first
		err := cache.Set(key, value, &expireTime)
		assert.NoError(t, err)

		// Verify it exists
		var result string
		err = cache.Get(key, &result)
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// Delete it
		err = cache.Delete(key)
		assert.NoError(t, err)

		// Verify it's gone
		var deleted string
		err = cache.Get(key, &deleted)
		assert.Error(t, err)
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("ClearWithPrefix", func(t *testing.T) {
		prefix := "test-prefix"
		expireTime := 5 * time.Minute

		// Set multiple keys with prefix
		keys := []string{"test-prefix:1", "test-prefix:2", "other-key"}
		for _, key := range keys {
			err := cache.Set(key, "value", &expireTime)
			assert.NoError(t, err)
		}

		// Clear with prefix
		err := cache.ClearWithPrefix(prefix)
		assert.NoError(t, err)

		// Verify prefixed keys are gone
		for _, key := range []string{"test-prefix:1", "test-prefix:2"} {
			var result string
			err := cache.Get(key, &result)
			assert.Error(t, err)
			assert.Equal(t, redis.Nil, err)
		}

		// Verify other key still exists
		var result string
		err = cache.Get("other-key", &result)
		assert.NoError(t, err)
		assert.Equal(t, "value", result)
	})

	t.Run("AddScore and GetTopN", func(t *testing.T) {
		boardKey := "test-leaderboard"

		// Add scores
		err := cache.AddScore(boardKey, "player1", 100.0)
		assert.NoError(t, err)

		err = cache.AddScore(boardKey, "player2", 200.0)
		assert.NoError(t, err)

		err = cache.AddScore(boardKey, "player3", 150.0)
		assert.NoError(t, err)

		// Get top 3
		topN, err := cache.GetTopN(boardKey, 3)
		assert.NoError(t, err)
		assert.Len(t, topN, 3)

		// Verify order (should be descending by score)
		assert.Equal(t, "player2", topN[0].Member)
		assert.Equal(t, 200.0, topN[0].Score)
		assert.Equal(t, "player3", topN[1].Member)
		assert.Equal(t, 150.0, topN[1].Score)
		assert.Equal(t, "player1", topN[2].Member)
		assert.Equal(t, 100.0, topN[2].Score)
	})

	t.Run("GetRank", func(t *testing.T) {
		boardKey := "test-leaderboard"

		// Get rank for player2 (should be 1st)
		rank, score, err := cache.GetRank(boardKey, "player2")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), rank)
		assert.Equal(t, 200.0, score)

		// Get rank for player1 (should be 3rd)
		rank, score, err = cache.GetRank(boardKey, "player1")
		assert.NoError(t, err)
		assert.Equal(t, int64(3), rank)
		assert.Equal(t, 100.0, score)
	})

	t.Run("RemoveMember", func(t *testing.T) {
		boardKey := "test-leaderboard"

		// Remove player2
		err := cache.RemoveMember(boardKey, "player2")
		assert.NoError(t, err)

		// Verify player2 is gone
		_, _, err = cache.GetRank(boardKey, "player2")
		assert.Error(t, err)

		// Verify other players are still there
		rank, _, err := cache.GetRank(boardKey, "player1")
		assert.NoError(t, err)
		assert.Equal(t, int64(2), rank) // Should now be 2nd instead of 3rd
	})

	t.Run("GetAroundMember", func(t *testing.T) {
		boardKey := "test-leaderboard"

		// Get around player3 with radius 1
		around, err := cache.GetAroundMember(boardKey, "player3", 1)
		assert.NoError(t, err)
		assert.Len(t, around, 2) // player3 and player1

		// Verify the order
		assert.Equal(t, "player3", around[0].Member)
		assert.Equal(t, "player1", around[1].Member)
	})

	t.Run("Clear", func(t *testing.T) {
		// Clear all data
		err := cache.Clear()
		assert.NoError(t, err)

		// Verify everything is gone
		var result string
		err = cache.Get("other-key", &result)
		assert.Error(t, err)
		assert.Equal(t, redis.Nil, err)
	})
}

// Test error cases with mock Redis client
func TestAppCache_ErrorCases(t *testing.T) {
	t.Run("Set error", func(t *testing.T) {
		// This would require a more sophisticated mock
		// For now, we'll test the integration test covers the happy path
		t.Skip("Requires sophisticated Redis mocking")
	})
}

// MockRedisClient for error testing (simplified)
type MockRedisClient struct{}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx, "SET", key, value, "EX", int(expiration.Seconds()))
	cmd.SetErr(errors.New("mock redis error"))
	return cmd
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx, "GET", key)
	cmd.SetErr(errors.New("mock redis error"))
	return cmd
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := make([]interface{}, len(keys)+1)
	args[0] = "DEL"
	for i, key := range keys {
		args[i+1] = key
	}
	cmd := redis.NewIntCmd(ctx, args...)
	cmd.SetErr(errors.New("mock redis error"))
	return cmd
}

func (m *MockRedisClient) FlushAll(ctx context.Context) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx, "FLUSHALL")
	cmd.SetErr(errors.New("mock redis error"))
	return cmd
}

func (m *MockRedisClient) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	cmd := redis.NewStringSliceCmd(ctx, "KEYS", pattern)
	cmd.SetErr(errors.New("mock redis error"))
	return cmd
}

func (m *MockRedisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx, "ZADD", key, members)
	cmd.SetErr(errors.New("mock redis error"))
	return cmd
}

func (m *MockRedisClient) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	cmd := redis.NewZSliceCmd(ctx, "ZREVRANGE", key, start, stop, "WITHSCORES")
	cmd.SetErr(errors.New("mock redis error"))
	return cmd
}

func (m *MockRedisClient) ZRevRank(ctx context.Context, key, member string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx, "ZREVRANK", key, member)
	cmd.SetErr(errors.New("mock redis error"))
	return cmd
}

func (m *MockRedisClient) ZScore(ctx context.Context, key, member string) *redis.FloatCmd {
	cmd := redis.NewFloatCmd(ctx, "ZSCORE", key, member)
	cmd.SetErr(errors.New("mock redis error"))
	return cmd
}

func (m *MockRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	args := make([]interface{}, len(members)+2)
	args[0] = "ZREM"
	args[1] = key
	for i, member := range members {
		args[i+2] = member
	}
	cmd := redis.NewIntCmd(ctx, args...)
	cmd.SetErr(errors.New("mock redis error"))
	return cmd
}

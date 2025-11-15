package rstream

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/hiamthach108/simplerank/config"
	"github.com/hiamthach108/simplerank/internal/service"
	"github.com/hiamthach108/simplerank/pkg/cache"
	"github.com/hiamthach108/simplerank/pkg/logger"
	"go.uber.org/fx"
)

type Subscriber struct {
	config     *config.AppConfig
	cache      cache.ICache
	logger     logger.ILogger
	historySvc service.IHistorySvc
}

func NewSubscriber(
	config *config.AppConfig,
	cache cache.ICache,
	logger logger.ILogger,
	historySvc service.IHistorySvc,
) *Subscriber {
	return &Subscriber{
		config:     config,
		cache:      cache,
		logger:     logger,
		historySvc: historySvc,
	}
}

// Start initializes and starts all stream subscriptions
func (s *Subscriber) Start(ctx context.Context) error {
	s.logger.Info("Starting Redis Stream Subscriber...")

	// Subscribe here
	if err := s.subscribeToLeaderboardUpdates(ctx); err != nil {
		return err
	}

	s.logger.Info("Redis Stream Subscriber started successfully")
	return nil
}

// Stop gracefully shuts down the subscriber
func (s *Subscriber) Stop(ctx context.Context) error {
	s.logger.Info("Stopping Redis Stream Subscriber...")
	// Add any cleanup logic here if needed
	return nil
}

// RegisterHooks registers the subscriber lifecycle hooks with fx
func RegisterHooks(lc fx.Lifecycle, subscriber *Subscriber) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return subscriber.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return subscriber.Stop(ctx)
		},
	})
}

// decodeMessage is a generic helper function to decode gob-encoded messages
func decodeMessage[T any](message any) (*T, error) {
	// Extract binary data from message
	messageMap, ok := message.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("failed to cast message to map")
	}

	data, ok := messageMap["data"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to extract data field from message")
	}

	// Decode using gob
	var result T
	buf := bytes.NewBufferString(data)
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode message: %w", err)
	}

	return &result, nil
}

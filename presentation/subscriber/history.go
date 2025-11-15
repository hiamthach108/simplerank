package subscriber

import (
	"context"

	"github.com/hiamthach108/simplerank/internal/dto"
	"github.com/hiamthach108/simplerank/internal/shared/constants"
	"github.com/hiamthach108/simplerank/pkg/cache"
)

// subscribeToLeaderboardUpdates subscribes to leaderboard updates stream
func (s *Subscriber) subscribeToLeaderboardUpdates(ctx context.Context) error {
	stream := constants.STREAM_LEADERBOARD_UPDATE
	group := constants.STREAM_LEADERBOARD_GROUP

	// Ensure consumer group exists
	if err := s.cache.EnsureGroup(stream, group); err != nil {
		s.logger.Error("Failed to ensure consumer group", "stream", stream, "group", group, "error", err)
		return err
	}

	// Create handler
	handler := cache.ConsumerHandler{
		Consumer: constants.STREAM_HISTORY_CONSUMER_GROUP,
		Handler: func(message any) {
			// Process leaderboard update message
			s.logger.Info("Received leaderboard update", "message", message)

			// Decode message using generic helper
			req, err := decodeMessage[dto.CreateHistoryReq](message)
			if err != nil {
				s.logger.Error("Failed to decode message", "error", err)
				return
			}

			history, err := s.historySvc.Record(ctx, req)
			if err != nil {
				s.logger.Error("Failed to record history", "error", err)
				return
			}
			s.logger.Info("Recorded history successfully", "history", history)
		},
	}

	// Subscribe to stream
	if err := s.cache.Subscribe(stream, group, handler); err != nil {
		s.logger.Error("Failed to subscribe to stream", "stream", stream, "error", err)
		return err
	}

	s.logger.Info("Subscribed to stream", "stream", stream, "group", group, "consumer", handler.Consumer)
	return nil
}

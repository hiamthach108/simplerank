package main

import (
	"github.com/hiamthach108/simplerank/config"
	"github.com/hiamthach108/simplerank/internal/repository"
	"github.com/hiamthach108/simplerank/internal/service"
	"github.com/hiamthach108/simplerank/pkg/cache"
	"github.com/hiamthach108/simplerank/pkg/database"

	"github.com/hiamthach108/simplerank/pkg/logger"
	"github.com/hiamthach108/simplerank/presentation/http"
	"github.com/hiamthach108/simplerank/presentation/rstream"
	"github.com/hiamthach108/simplerank/presentation/socket"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Provide(
			// Core
			config.NewAppConfig,
			logger.NewLogger,
			cache.NewAppCache,
			database.NewDbClient,
			database.NewClickHouseDbClient,
			http.NewHttpServer,
			rstream.NewSubscriber,
			socket.NewHub,
			fx.Annotate(
				func(hub *socket.Hub) socket.IBroadcaster {
					return hub
				},
				fx.As(new(socket.IBroadcaster)),
			),

			// Services
			service.NewLeaderBoardSvc,
			service.NewHistorySvc,

			// Repositories
			repository.NewLeaderboardRepository,
			repository.NewHistoryRepository,
		),
		fx.Invoke(http.RegisterHooks),
		fx.Invoke(rstream.RegisterHooks),
		fx.Invoke(socket.RegisterHooks),
	)

	app.Run()
}

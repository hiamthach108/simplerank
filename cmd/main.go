package main

import (
	"github.com/hiamthach108/simplerank/config"
	"github.com/hiamthach108/simplerank/internal/repository"
	"github.com/hiamthach108/simplerank/internal/service"
	"github.com/hiamthach108/simplerank/pkg/cache"
	"github.com/hiamthach108/simplerank/pkg/database"

	"github.com/hiamthach108/simplerank/pkg/logger"
	"github.com/hiamthach108/simplerank/presentation/http"
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

			// Services
			service.NewLeaderBoardSvc,
			service.NewHistorySvc,

			// Repositories
			repository.NewLeaderboardRepository,
			repository.NewHistoryRepository,
		),
		fx.Invoke(http.RegisterHooks),
	)

	app.Run()
}

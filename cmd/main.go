package main

import (
	"github.com/hiamthach108/simplerank/config"
	"github.com/hiamthach108/simplerank/pkg/cache"
	"github.com/hiamthach108/simplerank/pkg/logger"
	"github.com/hiamthach108/simplerank/presentation/http"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Provide(
			config.NewAppConfig,
			logger.NewLogger,
			cache.NewAppCache,
			http.NewHttpServer,
		),
		fx.Invoke(http.RegisterHooks),
	)

	app.Run()
}

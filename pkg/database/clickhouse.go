package database

import (
	"context"
	"fmt"
	"log"

	"github.com/hiamthach108/simplerank/config"
	"github.com/hiamthach108/simplerank/internal/model"
	"github.com/hiamthach108/simplerank/pkg/logger"
	"github.com/uptrace/go-clickhouse/ch"
)

func NewClickHouseDbClient(config *config.AppConfig, logger logger.ILogger) (*ch.DB, error) {
	db := ch.Connect(
		ch.WithAddr(fmt.Sprintf("%s:%d", config.ClickHouse.Host, config.ClickHouse.Port)),
		ch.WithDatabase(config.ClickHouse.DB),
		ch.WithUser(config.ClickHouse.User),
		ch.WithPassword(config.ClickHouse.Password),
		ch.WithCompression(true),
		ch.WithAutoCreateDatabase(true),
	)
	defer db.Close()

	ctx := context.Background()
	if err := createTables(db, ctx); err != nil {
		log.Fatal(err)
		return nil, err
	}

	return db, nil
}

func createTables(db *ch.DB, ctx context.Context) error {
	models := []any{
		(*model.History)(nil),
	}

	for _, model := range models {
		_, err := db.NewCreateTable().
			Model(model).
			Engine("MergeTree()").
			Order("id").
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create table for model %T: %w", model, err)
		}
	}

	return nil
}

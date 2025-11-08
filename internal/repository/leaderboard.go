package repository

import (
	"github.com/hiamthach108/simplerank/internal/model"
	"gorm.io/gorm"
)

type ILeaderboardRepository interface {
	IRepository[model.Leaderboard]
}

type leaderboardRepository struct {
	Repository[model.Leaderboard]
}

func NewLeaderboardRepository(dbClient *gorm.DB) ILeaderboardRepository {
	return &leaderboardRepository{
		Repository: Repository[model.Leaderboard]{dbClient: dbClient},
	}
}

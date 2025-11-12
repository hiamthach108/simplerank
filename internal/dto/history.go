package dto

import (
	"encoding/json"
	"time"

	"github.com/hiamthach108/simplerank/internal/model"
	"gorm.io/datatypes"
)

type HistoryDto struct {
	LeaderboardID string    `json:"leaderboardId"`
	EntryID       string    `json:"entryId"`
	Score         float64   `json:"score"`
	CreatedAt     time.Time `json:"createdAt"`
	Metadata      any       `json:"metadata,omitempty"`
}

func (HistoryDto) FromModel(m *model.History) HistoryDto {
	return HistoryDto{
		LeaderboardID: m.LeaderboardID,
		EntryID:       m.EntryID,
		Score:         m.Score,
		CreatedAt:     m.CreatedAt,
		Metadata:      m.Metadata,
	}
}

type CreateHistoryReq struct {
	LeaderboardID string  `json:"leaderboardId" binding:"required"`
	EntryID       string  `json:"entryId" binding:"required"`
	Score         float64 `json:"score" binding:"required"`
	Metadata      any     `json:"metadata"`
}

func (r *CreateHistoryReq) ToModel() *model.History {
	m := &model.History{
		LeaderboardID: r.LeaderboardID,
		EntryID:       r.EntryID,
		Score:         r.Score,
	}
	if r.Metadata != nil {
		if data, err := json.Marshal(r.Metadata); err == nil {
			m.Metadata = datatypes.JSON(data)
		}
	}

	return m
}

type ListHistoriesReq struct {
	PaginationReq
	LeaderboardID string     `form:"leaderboardId" binding:"required"`
	EntryID       *string    `form:"entryId"`
	FromScore     *float64   `form:"fromScore"`
	ToScore       *float64   `form:"toScore"`
	FromDate      *time.Time `form:"fromDate"`
	ToDate        *time.Time `form:"toDate"`
}

package dto

import (
	"time"

	"github.com/hiamthach108/simplerank/internal/model"
)

type UpdateEntryScore struct {
	EntryID string  `json:"entryId" binding:"required"`
	Score   float64 `json:"score" binding:"required"`
}

type CreateLeaderboardReq struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	ExpiredAt   time.Time `json:"expiredAt" binding:"required"`
}

type LeaderboardDto struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ExpiredAt   time.Time `json:"expiredAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	TopEntries  any       `json:"topEntries,omitempty"`
}

func (d *LeaderboardDto) ToModel() *model.Leaderboard {
	return &model.Leaderboard{
		BaseModel: model.BaseModel{
			ID:        d.ID,
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
		},
		Name:        d.Name,
		Description: d.Description,
		ExpiredAt:   d.ExpiredAt,
	}
}

func (d *LeaderboardDto) FromModel(m *model.Leaderboard) {
	d.ID = m.ID
	d.Name = m.Name
	d.Description = m.Description
	d.ExpiredAt = m.ExpiredAt
	d.CreatedAt = m.CreatedAt
	d.UpdatedAt = m.UpdatedAt
}

type UpdateLeaderboardReq struct {
	ID          string     `json:"id"`
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	ExpiredAt   *time.Time `json:"expiredAt"`
}

func (r *UpdateLeaderboardReq) ToModel() (u *model.Leaderboard, fields []string) {
	u = &model.Leaderboard{}
	if r.Name != nil {
		u.Name = *r.Name
		fields = append(fields, "name")
	}
	if r.Description != nil {
		u.Description = *r.Description
		fields = append(fields, "description")
	}
	if r.ExpiredAt != nil {
		u.ExpiredAt = *r.ExpiredAt
		fields = append(fields, "expired_at")
	}
	return u, fields
}

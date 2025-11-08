package model

import "time"

type Leaderboard struct {
	BaseModel
	Name        string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
	ExpiredAt   time.Time `gorm:"not null;index"`
}

func (Leaderboard) TableName() string {
	return "leaderboards"
}

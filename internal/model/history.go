package model

type History struct {
	BaseModel
	LeaderboardID string  `gorm:"type:varchar(36);index"`
	EntryID       string  `gorm:"type:varchar(36);index"`
	Score         float64 `gorm:"type:double precision"`
}

func (History) TableName() string {
	return "histories"
}

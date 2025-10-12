package dto

type UpdateEntryScore struct {
	EntryId string  `json:"entry_id"`
	Score   float64 `json:"score"`
}

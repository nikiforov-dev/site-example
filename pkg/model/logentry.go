package model

import "time"

type LogEntry struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `json:"user_id"`
	User        string     `json:"user"`
	Activity    string     `json:"activity"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Duration    int64      `json:"duration"` // в секундах
	Description string     `json:"description"`
}

package model

import "time"

// Простая модель сообщения
type Message struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	UserID    uint   `json:"user_id"`
	User      string `json:"user"`
	Text      string `json:"text"`
	CreatedAt time.Time
}

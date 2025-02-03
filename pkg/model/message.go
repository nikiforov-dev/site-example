package model

import "time"

// Простая модель сообщения
type Message struct {
	User      string `json:"user"`
	Text      string `json:"text"`
	CreatedAt time.Time
}

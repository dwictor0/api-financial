package models

import (
	"time"
)

type WebhookEvent struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	EventID      string    `gorm:"uniqueIndex;not null" json:"event_id"`
	CardID       string    `gorm:"not null" json:"card_id"`
	ClienteEmail string    `gorm:"not null" json:"cliente_email"`
	Timestamp    time.Time `json:"timestamp"`
	CreatedAt    time.Time `json:"created_at"`
}

type WebhookCardUpdatedInput struct {
	EventID      string `json:"event_id" binding:"required"`
	CardID       string `json:"card_id" binding:"required"`
	ClienteEmail string `json:"cliente_email" binding:"required,email"`
	Timestamp    string `json:"timestamp" binding:"required"`
}

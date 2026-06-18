package entity

import (
	"github.com/google/uuid"
	"time"
)

type Message struct {
	UUID       uuid.UUID `json:"uuid"`
	ChatUUID   uuid.UUID `json:"chat_uuid"`
	SenderUUID uuid.UUID `json:"sender_uuid"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Deleted    bool      `json:"deleted"`
}

package entity

import (
	"github.com/google/uuid"
	"time"
)

type Chat struct {
	UUID      uuid.UUID `json:"uuid"`
	Name      *string   `json:"name,omitempty"`
	Type      string    `json:"type"` // "personal", "group"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatParticipant struct {
	ChatUUID uuid.UUID `json:"chat_uuid"`
	UserUUID uuid.UUID `json:"user_uuid"`
	JoinedAt time.Time `json:"joined_at"`
}

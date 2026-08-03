package entity

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	UUID       uuid.UUID  `json:"uuid"`
	ChatUUID   uuid.UUID  `json:"chat_uuid"`
	SenderUUID uuid.UUID  `json:"sender_uuid"`
	Body       string     `json:"body"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Deleted    bool       `json:"deleted"`
	Edited     bool       `json:"edited"`
	ReplyTo    *uuid.UUID `json:"reply_to,omitempty"`
}

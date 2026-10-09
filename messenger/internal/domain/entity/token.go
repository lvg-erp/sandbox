package entity

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	UUID      uuid.UUID `json:"-"`
	Token     string    `json:"-"`
	UserUUID  uuid.UUID `json:"-"`
	CreatedAt time.Time `json:"-"`
	ExpiresAt time.Time `json:"-"`
	Revoked   bool      `json:"-"`
}

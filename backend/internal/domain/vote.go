package domain

import (
	"time"

	"github.com/google/uuid"
)

type Vote struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null"`
	User      User      `gorm:"constraint:OnDelete:CASCADE"`
	VideoID   uuid.UUID `gorm:"type:uuid;index;not null"`
	Video     Video     `gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}


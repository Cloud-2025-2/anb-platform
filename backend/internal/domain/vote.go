package domain

import (
	"time"

	"github.com/google/uuid"
)

type Vote struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null"`
	User      User      `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	VideoID   uuid.UUID `gorm:"type:uuid;index;not null"`
	Video     Video     `gorm:"foreignKey:VideoID;references:ID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}


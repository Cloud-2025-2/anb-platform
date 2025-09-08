package domain

import (
	"time"

	"github.com/google/uuid"
)


type VideoStatus string

const (
	VideoUploaded   VideoStatus = "uploaded"
	VideoProcessing VideoStatus = "processing"
	VideoProcessed  VideoStatus = "processed"
	VideoFailed     VideoStatus = "failed"
)

type Video struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID          uuid.UUID   `gorm:"type:uuid;index;not null"`
	User            User        `gorm:"constraint:OnDelete:CASCADE"`
	Title           string      `gorm:"not null"`
	OriginalURL     string      `gorm:"not null"`     // ruta/URL del archivo subido
	ProcessedURL    *string                              // se llena cuando termina worker
	Status          VideoStatus `gorm:"type:text;index;not null;default:uploaded"`
	UploadedAt      time.Time   `gorm:"autoCreateTime"`
	ProcessedAt     *time.Time
	DurationOrigSec *int
	DurationProcSec *int
	WidthOrig       *int
	HeightOrig      *int
	WidthProc       *int
	HeightProc      *int
	AspectProc      *string // "16:9"
	HasAudioOrig    *bool
	Watermark       bool      `gorm:"default:false"`
	IsPublicForVote bool      `gorm:"default:false;index"`
	CitySnapshot    string    // copia de city del usuario para ranking por ciudad
	ChecksumSHA256  *string

	Votes []Vote `gorm:"foreignKey:VideoID"`
}
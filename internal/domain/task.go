package domain

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskQueued    TaskStatus = "queued"
	TaskRunning   TaskStatus = "running"
	TaskSucceeded TaskStatus = "succeeded"
	TaskFailed    TaskStatus = "failed"
	TaskRetrying  TaskStatus = "retrying"
	TaskDead      TaskStatus = "dead"
)

type ProcessingTask struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	VideoID     uuid.UUID  `gorm:"type:uuid;index;not null"`
	Video       Video      `gorm:"constraint:OnDelete:CASCADE"`
	TaskType    string     `gorm:"not null"` // p.ej. "video:process"
	Status      TaskStatus `gorm:"type:text;not null;default:queued"`
	Attempts    int        `gorm:"not null;default:0"`
	MaxAttempts int        `gorm:"not null;default:5"`
	LastError   *string
	EnqueuedAt  time.Time  `gorm:"autoCreateTime"`
	StartedAt   *time.Time
	FinishedAt  *time.Time
	WorkerID    *string
}

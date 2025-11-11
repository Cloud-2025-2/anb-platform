package queue

import (
	"context"
	"time"
)

type VideoProcessingTask struct {
	VideoID    string    `json:"video_id"`
	UserID     string    `json:"user_id"`
	Title      string    `json:"title"`
	FilePath   string    `json:"file_path"` // S3 key o URL
	Timestamp  time.Time `json:"timestamp"`
	RetryCount int       `json:"retry_count"`
}

type Producer interface {
	PublishVideoProcessingTask(t VideoProcessingTask) error
	Close() error
}

type Consumer interface {
	Start(ctx context.Context, handle func(VideoProcessingTask) error) error
	Close() error
}

package video

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/Cloud-2025-2/anb-platform/internal/domain"
	"github.com/Cloud-2025-2/anb-platform/internal/kafka"
	"github.com/Cloud-2025-2/anb-platform/internal/repo"
)

type Storage interface {
	// Guarda un archivo temporal en storage
	Save(localTmpPath, destName string) (string, error)
}

type Service struct {
	videos   repo.VideoRepository
	store    Storage
	producer *kafka.Producer
}

func NewService(videos repo.VideoRepository, store Storage, producer *kafka.Producer) *Service {
	return &Service{videos: videos, store: store, producer: producer}
}

// UploadAndEnqueue guarda metadata del video y crea una tarea asíncrona
func (s *Service) UploadAndEnqueue(user domain.User, tmpPath, title string) (taskID string, videoID uuid.UUID, err error) {
	// 1. Guardar archivo en storage
	destName := uuid.New().String() + filepath.Ext(tmpPath)
	storedPath, err := s.store.Save(tmpPath, destName)
	if err != nil {
		return "", uuid.Nil, err
	}

	// 2. Generate proper URL (S3 or local)
	s3Bucket := os.Getenv("S3_BUCKET")
	awsRegion := os.Getenv("AWS_REGION")
	var url string
	if s3Bucket != "" && awsRegion != "" {
		// S3 URL format
		url = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s3Bucket, awsRegion, storedPath)
	} else {
		// Local storage URL
		url = fmt.Sprintf("/storage/%s", storedPath)
	}

	// 3. Crear registro en la DB
	v := domain.Video{
		UserID:       user.ID,
		Title:        title,
		OriginalURL:  url,
		Status:       domain.VideoUploaded,
		CitySnapshot: user.City,
	}
	if err := s.videos.Create(&v); err != nil {
		return "", uuid.Nil, err
	}

	// 4. Encolar tarea para el worker usando Kafka
	// Use the stored path (S3 key) for the worker, not the full URL
	task := kafka.VideoProcessingTask{
		VideoID:    v.ID.String(),
		UserID:     user.ID.String(),
		Title:      title,
		FilePath:   storedPath,
		Timestamp:  time.Now(),
		RetryCount: 0,
	}

	if err := s.producer.PublishVideoProcessingTask(task); err != nil {
		return "", uuid.Nil, err
	}

	return uuid.New().String(), v.ID, nil
}

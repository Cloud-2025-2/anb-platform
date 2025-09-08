package video

import (
	"encoding/json"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/Cloud-2025-2/anb-platform/internal/domain"
	"github.com/Cloud-2025-2/anb-platform/internal/repo"
)

type Storage interface {
	// Guarda un archivo temporal en storage 
	Save(localTmpPath, destName string) (string, error)
}

type Service struct {
	videos repo.VideoRepository
	store  Storage
	q      *asynq.Client
}

func NewService(videos repo.VideoRepository, store Storage, q *asynq.Client) *Service {
	return &Service{videos: videos, store: store, q: q}
}

// UploadAndEnqueue guarda metadata del video y crea una tarea asíncrona
func (s *Service) UploadAndEnqueue(user domain.User, tmpPath, title string) (taskID string, videoID uuid.UUID, err error) {
	// 1. Guardar archivo en storage
	destName := uuid.New().String() + filepath.Ext(tmpPath)
	url, err := s.store.Save(tmpPath, destName)
	if err != nil {
		return "", uuid.Nil, err
	}

	// 2. Crear registro en la DB
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

	// 3. Encolar tarea para el worker
	payload := map[string]string{"video_id": v.ID.String()}
	b, _ := json.Marshal(payload)
	task := asynq.NewTask("video:process", b)

	ti, err := s.q.Enqueue(task, asynq.Queue("default"))
	if err != nil {
		return "", uuid.Nil, err
	}

	return ti.ID, v.ID, nil
}

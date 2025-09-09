package repo

import (
	"github.com/Cloud-2025-2/anb-platform/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VideoRepository interface {
	Create(v *domain.Video) error
	FindByUser(userID uuid.UUID) ([]domain.Video, error)
	FindByIDForUser(id, userID uuid.UUID) (*domain.Video, error)
	FindByID(id uuid.UUID) (*domain.Video, error)                   // útil para el worker
	Update(v *domain.Video) error
	DeleteByIDForUser(id, userID uuid.UUID) error                   // usado por handler Delete
	ListPublic(limit, offset int) ([]domain.Video, error)           // usado por público
}

type videoRepo struct{ db *gorm.DB }

func NewVideoRepo(db *gorm.DB) VideoRepository { return &videoRepo{db} }

func (r *videoRepo) Create(v *domain.Video) error {
	return r.db.Create(v).Error
}

func (r *videoRepo) FindByUser(userID uuid.UUID) ([]domain.Video, error) {
	var out []domain.Video
	err := r.db.Where("user_id = ?", userID).
		Order("uploaded_at DESC").
		Find(&out).Error
	return out, err
}

func (r *videoRepo) FindByIDForUser(id, userID uuid.UUID) (*domain.Video, error) {
	var v domain.Video
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *videoRepo) FindByID(id uuid.UUID) (*domain.Video, error) {
	var v domain.Video
	if err := r.db.First(&v, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *videoRepo) Update(v *domain.Video) error {
	return r.db.Save(v).Error
}

func (r *videoRepo) DeleteByIDForUser(id, userID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// First verify the video exists and belongs to the user
		var video domain.Video
		if err := tx.Where("id = ? AND user_id = ?", id, userID).First(&video).Error; err != nil {
			return err // This will return gorm.ErrRecordNotFound if not found
		}
		
		// Delete all votes for this video first (cascade prevention)
		if err := tx.Where("video_id = ?", id).Delete(&domain.Vote{}).Error; err != nil {
			return err
		}
		
		// Note: Tasks are handled by the async processing system separately
		
		// Finally delete the video
		if err := tx.Delete(&video).Error; err != nil {
			return err
		}
		
		return nil
	})
}

func (r *videoRepo) ListPublic(limit, offset int) ([]domain.Video, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	var out []domain.Video
	err := r.db.Where("status = ? AND is_public_for_vote = ?", "published", true).
		Order("uploaded_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&out).Error
	return out, err
}

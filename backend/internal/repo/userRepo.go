package repo

import (
	"github.com/Cloud-2025-2/anb-platform/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(u *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByID(id uuid.UUID) (*domain.User, error)
	DeleteByID(id uuid.UUID) error
}

type userRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) UserRepository { return &userRepo{db} }

func (r *userRepo) Create(u *domain.User) error { return r.db.Create(u).Error }

func (r *userRepo) FindByEmail(email string) (*domain.User, error) {
	var u domain.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil { return nil, err }
	return &u, nil
}

func (r *userRepo) FindByID(id uuid.UUID) (*domain.User, error) {
	var u domain.User
	if err := r.db.First(&u, "id = ?", id).Error; err != nil { return nil, err }
	return &u, nil
}

func (r *userRepo) DeleteByID(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// First verify user exists
		var user domain.User
		if err := tx.Where("id = ?", id).First(&user).Error; err != nil {
			return err
		}
		
		// Get all user's videos to delete votes on them
		var userVideos []domain.Video
		if err := tx.Where("user_id = ?", id).Find(&userVideos).Error; err != nil {
			return err
		}
		
		// Delete all votes by this user
		if err := tx.Where("user_id = ?", id).Delete(&domain.Vote{}).Error; err != nil {
			return err
		}
		
		// Delete all votes on user's videos
		for _, video := range userVideos {
			if err := tx.Where("video_id = ?", video.ID).Delete(&domain.Vote{}).Error; err != nil {
				return err
			}
		}
		
		// Delete user's videos
		if err := tx.Where("user_id = ?", id).Delete(&domain.Video{}).Error; err != nil {
			return err
		}
		
		// Finally delete the user
		return tx.Delete(&user).Error
	})
}

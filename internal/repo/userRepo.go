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

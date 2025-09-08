package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RolePlayer Role = "player"
	RoleJury   Role = "jury"
	RoleAdmin  Role = "admin"
	RolePublic Role = "public"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FirstName    string    `gorm:"not null"`
	LastName     string    `gorm:"not null"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"` // solo 1 hash, password2 no se almacena
	City         string    `gorm:"not null"`
	Country      string    `gorm:"not null"`
	Role         Role      `gorm:"type:text;not null;default:player"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Videos []Video `gorm:"foreignKey:UserID"`}
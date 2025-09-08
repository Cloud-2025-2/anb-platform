package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/Cloud-2025-2/anb-platform/internal/domain"
	"github.com/Cloud-2025-2/anb-platform/internal/repo"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	users   repo.UserRepository
	secret  string
	expires time.Duration
}

func NewService(users repo.UserRepository, secret string, expiresMinutes int) *Service {
	return &Service{users: users, secret: secret, expires: time.Duration(expiresMinutes) * time.Minute}
}

func (s *Service) SignUp(in domain.User, password1, password2 string) error {
	if password1 != password2 {
		return errors.New("passwords do not match")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(password1), bcrypt.DefaultCost)
	in.PasswordHash = string(hash)
	in.Role = "player"
	return s.users.Create(&in)
}

type LoginResult struct {
	Token     string `json:"access_token"`
	TokenType string `json:"token_type"`
	ExpiresIn int    `json:"expires_in"`
}

func (s *Service) Login(email, password string) (*LoginResult, error) {
	u, err := s.users.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, errors.New("invalid credentials")
	}
	claims := jwt.MapClaims{
		"sub": u.ID.String(),
		"exp": time.Now().Add(s.expires).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := tok.SignedString([]byte(s.secret))
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token:     tokenStr,
		TokenType: "Bearer",
		ExpiresIn: int(s.expires.Seconds()),
	}, nil
}

func (s *Service) DeleteUser(userID uuid.UUID) error {
	return s.users.DeleteByID(userID)
}

package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Cloud-2025-2/anb-platform/internal/auth"
	"github.com/Cloud-2025-2/anb-platform/internal/domain"
)

type AuthHandlers struct{ svc *auth.Service }

func NewAuthHandlers(s *auth.Service) *AuthHandlers { return &AuthHandlers{svc: s} }

func (h *AuthHandlers) SignUp(c *gin.Context) {
	var in SignUpIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u := domain.User{
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Email:     in.Email,
		City:      in.City,
		Country:   in.Country,
	}
	if err := h.svc.SignUp(u, in.Password1, in.Password2); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *AuthHandlers) Login(c *gin.Context) {
	var in LoginIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := h.svc.Login(in.Email, in.Password)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
	})
}

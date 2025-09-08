package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Cloud-2025-2/anb-platform/internal/auth"
	"github.com/Cloud-2025-2/anb-platform/internal/domain"
)

type AuthHandlers struct{ svc *auth.Service }

func NewAuthHandlers(s *auth.Service) *AuthHandlers { return &AuthHandlers{svc: s} }

// SignUp godoc
// @Summary Register a new player
// @Description Register a new player in the ANB platform
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body SignUpIn true "User registration data"
// @Success 201 {object} map[string]string "User created successfully"
// @Failure 400 {object} map[string]string "Bad request - validation error"
// @Failure 409 {object} map[string]string "Conflict - email already exists"
// @Failure 422 {object} map[string]string "Unprocessable entity - password validation failed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/signup [post]
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
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// Login godoc
// @Summary User authentication
// @Description Authenticate user and generate JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginIn true "User login credentials"
// @Success 200 {object} auth.LoginResult "Authentication successful"
// @Failure 400 {object} map[string]string "Bad request - validation error"
// @Failure 401 {object} map[string]string "Unauthorized - invalid credentials"
// @Failure 404 {object} map[string]string "Not found - user does not exist"
// @Failure 429 {object} map[string]string "Too many requests - rate limit exceeded"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandlers) Login(c *gin.Context) {
	var in LoginIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.svc.Login(in.Email, in.Password)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, result)
}

// DeleteUser godoc
// @Summary Delete user account
// @Description Delete the authenticated user's account. This action is irreversible.
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "User deleted successfully"
// @Failure 401 {object} map[string]string "Unauthorized - invalid or missing token"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth [delete]
func (h *AuthHandlers) DeleteUser(c *gin.Context) {
	uid, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
		return
	}

	if err := h.svc.DeleteUser(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "User deleted successfully"})
}

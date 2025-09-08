package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Cloud-2025-2/anb-platform/internal/repo"
)

type PublicHandlers struct {
	videos repo.VideoRepository
	votes  repo.VoteRepository
	users  repo.UserRepository
}

func NewPublicHandlers(videos repo.VideoRepository, votes repo.VoteRepository, users repo.UserRepository) *PublicHandlers {
	return &PublicHandlers{videos: videos, votes: votes, users: users}
}

// GET /api/public/videos?limit=20&offset=0
func (h *PublicHandlers) ListVideos(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	list, err := h.videos.ListPublic(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// POST /api/public/videos/:id/vote  (JWT requerido; tu router ya lo montó bajo /api con middleware)
func (h *PublicHandlers) Vote(c *gin.Context) {
	uid, err := uuid.Parse(c.GetString("user_id"))
	if err != nil { c.Status(http.StatusUnauthorized); return }

	vid, err := uuid.Parse(c.Param("id"))
	if err != nil { c.Status(http.StatusNotFound); return }

	if err := h.votes.CastOnce(uid, vid); err != nil {
		// si UNIQUE(user_id,video_id) falla, significa que ya votó
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ya has votado por este video"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Voto registrado exitosamente."})
}

// GET /api/public/rankings?city=Bogota&limit=50
func (h *PublicHandlers) Rankings(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	city := c.DefaultQuery("city", "")
	rows, err := h.votes.TopByCity(limit, city)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

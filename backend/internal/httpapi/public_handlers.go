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

// ListVideos godoc
// @Summary List public videos
// @Description Get all videos available for public voting
// @Tags Public
// @Produce json
// @Param limit query int false "Number of videos to return (default: 20)"
// @Param offset query int false "Number of videos to skip (default: 0)"
// @Success 200 {array} domain.Video "List of public videos"
// @Failure 400 {object} map[string]string "Bad request - invalid query parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /public/videos [get]
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

// Vote godoc
// @Summary Vote for a video
// @Description Cast a vote for a public video. Requires authentication. One vote per user per video.
// @Tags Public
// @Produce json
// @Security BearerAuth
// @Param id path string true "Video ID"
// @Success 200 {object} map[string]string "Vote registered successfully"
// @Failure 400 {object} map[string]string "Bad request - already voted or invalid video ID"
// @Failure 401 {object} map[string]string "Unauthorized - invalid or missing token"
// @Failure 403 {object} map[string]string "Forbidden - cannot vote on own video"
// @Failure 404 {object} map[string]string "Video not found or not available for voting"
// @Failure 409 {object} map[string]string "Conflict - duplicate vote"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /public/videos/{id}/vote [post]
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

// Rankings godoc
// @Summary Get player rankings
// @Description Get current player rankings based on votes. Can be filtered by city.
// @Tags Public
// @Produce json
// @Param limit query int false "Number of rankings to return (default: 50)"
// @Param city query string false "Filter by city"
// @Success 200 {array} map[string]interface{} "Player rankings"
// @Failure 400 {object} map[string]string "Bad request - invalid query parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /public/rankings [get]
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

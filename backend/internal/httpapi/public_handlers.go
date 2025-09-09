package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Cloud-2025-2/anb-platform/internal/cache"
	"github.com/Cloud-2025-2/anb-platform/internal/domain"
	"github.com/Cloud-2025-2/anb-platform/internal/repo"
)

type PublicHandlers struct {
	videos repo.VideoRepository
	votes  repo.VoteRepository
	users  repo.UserRepository
	cache  *cache.RankingsCache
}

func NewPublicHandlers(videos repo.VideoRepository, votes repo.VoteRepository, users repo.UserRepository, cache *cache.RankingsCache) *PublicHandlers {
	return &PublicHandlers{videos: videos, votes: votes, users: users, cache: cache}
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
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"}); return }

	// Check if video exists and is public for voting
	video, err := h.videos.FindByID(vid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}
	if !video.IsPublicForVote || video.Status != domain.VideoPublished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video not available for voting"})
		return
	}

	if err := h.votes.CastOnce(uid, vid); err != nil {
		if errors.Is(err, repo.ErrDuplicateVote) {
			c.JSON(http.StatusConflict, gin.H{"error": "Ya has votado por este video"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	
	// Invalidate rankings cache since vote counts have changed
	ctx := context.Background()
	_ = h.cache.InvalidateAll(ctx) // Ignore errors to avoid blocking the response
	
	c.JSON(http.StatusOK, gin.H{"message": "Voto registrado exitosamente."})
}

// Rankings godoc
// @Summary Get player rankings
// @Description Get current player rankings based on votes. Can be filtered by city. Results are cached for improved performance.
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
	
	ctx := context.Background()
	
	// Try to get from cache first
	if cachedRankings, found := h.cache.GetRankings(ctx, limit, city); found {
		c.JSON(http.StatusOK, cachedRankings)
		return
	}
	
	// Cache miss - get from database
	rows, err := h.votes.TopByCity(limit, city)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Convert to cache format
	rankings := make([]cache.RankingResult, len(rows))
	for i, row := range rows {
		rankings[i] = cache.RankingResult{
			VideoID: row.VideoID,
			Votes:   row.Votes,
		}
	}
	
	// Store in cache (ignore errors to avoid blocking the response)
	_ = h.cache.SetRankings(ctx, limit, city, rankings)
	
	c.JSON(http.StatusOK, rows)
}

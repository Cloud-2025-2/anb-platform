package httpapi

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Cloud-2025-2/anb-platform/internal/domain"
	"github.com/Cloud-2025-2/anb-platform/internal/repo"
	vidsvc "github.com/Cloud-2025-2/anb-platform/internal/video"
)

type VideoHandlers struct {
	users  repo.UserRepository
	videos repo.VideoRepository
	svc    *vidsvc.Service
}

func NewVideoHandlers(users repo.UserRepository, videos repo.VideoRepository, svc *vidsvc.Service) *VideoHandlers {
	return &VideoHandlers{users: users, videos: videos, svc: svc}
}

// Upload godoc
// @Summary Upload a video
// @Description Upload a video file for processing. Requires authentication.
// @Tags Videos
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param video_file formData file true "Video file (MP4, max 100MB)"
// @Param title formData string true "Video title"
// @Success 201 {object} map[string]interface{} "Video uploaded successfully"
// @Failure 400 {object} map[string]string "Bad request - file validation error"
// @Failure 401 {object} map[string]string "Unauthorized - invalid or missing token"
// @Failure 403 {object} map[string]string "Forbidden - user not allowed to upload"
// @Failure 413 {object} map[string]string "Payload too large - file exceeds 100MB"
// @Failure 415 {object} map[string]string "Unsupported media type - invalid file format"
// @Failure 422 {object} map[string]string "Unprocessable entity - missing required fields"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /videos/upload [post]
func (h *VideoHandlers) Upload(c *gin.Context) {
	uid, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	u, err := h.users.FindByID(uid)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	title := c.PostForm("title")
	file, err := c.FormFile("video_file")
	if err != nil || file.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video_file required"})
		return
	}
	if file.Size > 100*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "max 100MB"})
		return
	}

	tmp := filepath.Join(os.TempDir(), "anb_"+uuid.NewString()+filepath.Ext(file.Filename))
	if err := c.SaveUploadedFile(file, tmp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer os.Remove(tmp)

	taskID, videoID, err := h.svc.UploadAndEnqueue(*u, tmp, title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message":  "Video subido correctamente. Procesamiento en curso.",
		"task_id":  taskID,
		"video_id": videoID,
	})
}

// MyVideos godoc
// @Summary List user's videos
// @Description Get all videos uploaded by the authenticated user
// @Tags Videos
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.Video "List of user's videos"
// @Failure 401 {object} map[string]string "Unauthorized - invalid or missing token"
// @Failure 403 {object} map[string]string "Forbidden - access denied"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /videos [get]
func (h *VideoHandlers) MyVideos(c *gin.Context) {
	uid, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	list, err := h.videos.FindByUser(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// Detail godoc
// @Summary Get video details
// @Description Get detailed information about a specific video owned by the user
// @Tags Videos
// @Produce json
// @Security BearerAuth
// @Param id path string true "Video ID"
// @Success 200 {object} domain.Video "Video details"
// @Failure 400 {object} map[string]string "Bad request - invalid video ID format"
// @Failure 401 {object} map[string]string "Unauthorized - invalid or missing token"
// @Failure 403 {object} map[string]string "Forbidden - video does not belong to user"
// @Failure 404 {object} map[string]string "Video not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /videos/{id} [get]
func (h *VideoHandlers) Detail(c *gin.Context) {
	uid, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	v, err := h.videos.FindByIDForUser(id, uid)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, v)
}

// Delete godoc
// @Summary Delete a video
// @Description Delete a video owned by the user (only if not published for voting)
// @Tags Videos
// @Produce json
// @Security BearerAuth
// @Param id path string true "Video ID"
// @Success 200 {object} map[string]interface{} "Video deleted successfully"
// @Failure 400 {object} map[string]string "Bad request - video cannot be deleted"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Video not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /videos/{id} [delete]
func (h *VideoHandlers) Delete(c *gin.Context) {
	uid, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	v, err := h.videos.FindByIDForUser(id, uid)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	// Check if video can be deleted based on status
	if v.Status == domain.VideoPublished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El video no puede ser eliminado (no está en estado \"uploaded\" o \"processed\" o ya está publicado)."})
		return
	}
	// Borrar archivos si tienes Storage.Delete (opcional: delega a servicio)
	// os.Remove(v.OriginalURL); if v.ProcessedURL != nil { os.Remove(*v.ProcessedURL) }

	// Usa Save/SoftDelete según tu repo; aquí reuse Update→Status, o implementa Delete en repo.
	if err := h.videos.DeleteByIDForUser(id, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "El video ha sido eliminado exitosamente.", "video_id": id})
}


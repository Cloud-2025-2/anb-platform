package httpapi

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Cloud-2025-2/anb-platform/internal/repo"
	vidsvc "github.com/Cloud-2025-2/anb-platform/internal/video")

type VideoHandlers struct {
	users  repo.UserRepository
	videos repo.VideoRepository
	svc    *vidsvc.Service
}

func NewVideoHandlers(users repo.UserRepository, videos repo.VideoRepository, svc *vidsvc.Service) *VideoHandlers {
	return &VideoHandlers{users: users, videos: videos, svc: svc}
}

func (h *VideoHandlers) Upload(c *gin.Context) {
	uid, err := uuid.Parse(c.GetString("user_id"))
	if err != nil { c.Status(http.StatusUnauthorized); return }
	u, err := h.users.FindByID(uid)
	if err != nil { c.Status(http.StatusUnauthorized); return }

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

func (h *VideoHandlers) MyVideos(c *gin.Context) {
	uid, err := uuid.Parse(c.GetString("user_id"))
	if err != nil { c.Status(http.StatusUnauthorized); return }
	list, err := h.videos.FindByUser(uid)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, list)
}

func (h *VideoHandlers) Detail(c *gin.Context) {
	uid, err := uuid.Parse(c.GetString("user_id"))
	if err != nil { c.Status(http.StatusUnauthorized); return }
	id, err := uuid.Parse(c.Param("id"))
	if err != nil { c.Status(http.StatusNotFound); return }
	v, err := h.videos.FindByIDForUser(id, uid)
	if err != nil { c.Status(http.StatusNotFound); return }
	c.JSON(http.StatusOK, v)
}

func (h *VideoHandlers) Delete(c *gin.Context) {
	uid, err := uuid.Parse(c.GetString("user_id"))
	if err != nil { c.Status(http.StatusUnauthorized); return }
	id, err := uuid.Parse(c.Param("id"))
	if err != nil { c.Status(http.StatusNotFound); return }

	v, err := h.videos.FindByIDForUser(id, uid)
	if err != nil { c.Status(http.StatusNotFound); return }

	// Regla simple: permitir borrar si no está publicado para votación
	if v.IsPublicForVote {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video is public for vote; cannot delete"})
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

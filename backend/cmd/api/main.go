package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"

	"github.com/Cloud-2025-2/anb-platform/internal/auth"
	"github.com/Cloud-2025-2/anb-platform/internal/config"
	"github.com/Cloud-2025-2/anb-platform/internal/db"
	"github.com/Cloud-2025-2/anb-platform/internal/domain"
	"github.com/Cloud-2025-2/anb-platform/internal/httpapi"
	"github.com/Cloud-2025-2/anb-platform/internal/repo"
	"github.com/Cloud-2025-2/anb-platform/internal/storage"
	videosvc "github.com/Cloud-2025-2/anb-platform/internal/video"
)

func main() {
	cfg := config.Load()

	// DB
	db.Connect() 
	_ = db.DB.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`).Error
	if err := db.DB.AutoMigrate(&domain.User{}, &domain.Video{}, &domain.Vote{}); err != nil {
		log.Fatal(err)
	}

	// repos
	usersRepo := repo.NewUserRepo(db.DB)
	videosRepo := repo.NewVideoRepo(db.DB)
	votesRepo := repo.NewVoteRepo(db.DB)

	// services
	authSvc := auth.NewService(usersRepo, cfg.JWTSecret, cfg.JWTExpireMinutes)
	queueCli := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	store := storage.NewLocal("./storage")
	videoSvc := videosvc.NewService(videosRepo, store, queueCli)

	// handlers
	authH := httpapi.NewAuthHandlers(authSvc)
	videoH := httpapi.NewVideoHandlers(usersRepo, videosRepo, videoSvc)
	publicH := httpapi.NewPublicHandlers(videosRepo, votesRepo, usersRepo)

	// router
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.String(200, "ok") })

	// Auth
	r.POST("/api/auth/signup", authH.SignUp)
	r.POST("/api/auth/login", authH.Login)

	// Privadas (JWT)
	api := r.Group("/api")
	api.Use(httpapi.JWT(cfg.JWTSecret))
	{
		api.POST("/videos/upload", videoH.Upload)
		api.GET("/videos", videoH.MyVideos)
		api.GET("/videos/:id", videoH.Detail)
		api.DELETE("/videos/:id", videoH.Delete)

		// votar requiere JWT (aunque sea /public)
		api.POST("/public/videos/:id/vote", publicH.Vote)
	}

	// Público sin auth
	r.GET("/api/public/videos", publicH.ListVideos)
	r.GET("/api/public/rankings", publicH.Rankings)

	log.Printf("API listening on :%s", cfg.AppPort)
	_ = r.Run(":" + cfg.AppPort)
}

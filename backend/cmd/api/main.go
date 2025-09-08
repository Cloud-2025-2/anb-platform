package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Cloud-2025-2/anb-platform/docs"
	"github.com/Cloud-2025-2/anb-platform/internal/auth"
	"github.com/Cloud-2025-2/anb-platform/internal/config"
	"github.com/Cloud-2025-2/anb-platform/internal/db"
	"github.com/Cloud-2025-2/anb-platform/internal/domain"
	"github.com/Cloud-2025-2/anb-platform/internal/httpapi"
	"github.com/Cloud-2025-2/anb-platform/internal/repo"
	"github.com/Cloud-2025-2/anb-platform/internal/storage"
	videosvc "github.com/Cloud-2025-2/anb-platform/internal/video"
)

// @title ANB Rising Stars Showcase API
// @version 1.0
// @description API REST escalable para la plataforma de descubrimiento de talento de baloncesto de la Asociación Nacional de Baloncesto (ANB)
// @termsOfService http://swagger.io/terms/

// @contact.name ANB Platform Team
// @contact.url http://www.anb.com/support
// @contact.email support@anb.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8000
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load environment variables from .env file
	_ = godotenv.Load()

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

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health godoc
	// @Summary Health check
	// @Description Check if the API is running
	// @Tags Health
	// @Produce plain
	// @Success 200 {string} string "ok"
	// @Router /health [get]
	r.GET("/api/health", func(c *gin.Context) { c.String(200, "ok") })

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

		// Eliminar usuario para uso en pruebas de Postman
		api.DELETE("/auth", authH.DeleteUser)
	}

	// Público sin auth
	r.GET("/api/public/videos", publicH.ListVideos)
	r.GET("/api/public/rankings", publicH.Rankings)

	log.Printf("API listening on :%s", cfg.AppPort)
	_ = r.Run(":" + cfg.AppPort)
}

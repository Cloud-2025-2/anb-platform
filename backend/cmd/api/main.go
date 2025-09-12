package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Cloud-2025-2/anb-platform/docs"
	"github.com/Cloud-2025-2/anb-platform/internal/auth"
	"github.com/Cloud-2025-2/anb-platform/internal/cache"
	"github.com/Cloud-2025-2/anb-platform/internal/config"
	"github.com/Cloud-2025-2/anb-platform/internal/db"
	"github.com/Cloud-2025-2/anb-platform/internal/domain"
	"github.com/Cloud-2025-2/anb-platform/internal/httpapi"
	"github.com/Cloud-2025-2/anb-platform/internal/kafka"
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

	// Kafka producer for video processing
	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// Redis client for caching
	redisCli := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0, // Use default DB for caching
	})

	// Initialize cache with 3-minute TTL (within the 1-5 minute range requested)
	rankingsCache := cache.NewRankingsCache(redisCli, 3*time.Minute)

	store := storage.NewLocal("./storage")
	videoSvc := videosvc.NewService(videosRepo, store, kafkaProducer)

	// handlers
	authH := httpapi.NewAuthHandlers(authSvc)
	videoH := httpapi.NewVideoHandlers(usersRepo, videosRepo, videoSvc)
	publicH := httpapi.NewPublicHandlers(videosRepo, votesRepo, usersRepo, rankingsCache)

	// router
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:3000",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
	r.GET("/api/public/cities", publicH.GetCities)

	log.Printf("API listening on :%s", cfg.AppPort)
	_ = r.Run(":" + cfg.AppPort)
}

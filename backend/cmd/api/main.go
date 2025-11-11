package main

import (
	"context"
	"log"
	"os"
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
	"github.com/Cloud-2025-2/anb-platform/internal/queue"
	sqsqueue "github.com/Cloud-2025-2/anb-platform/internal/queue/sqs"
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
	// 1) ENV
	_ = godotenv.Load()
	cfg := config.Load()

	// 2) DB & Migrations
	db.Connect()
	_ = db.DB.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`).Error
	if err := db.DB.AutoMigrate(&domain.User{}, &domain.Video{}, &domain.Vote{}); err != nil {
		log.Fatal(err)
	}

	// 3) Repos
	usersRepo := repo.NewUserRepo(db.DB)
	videosRepo := repo.NewVideoRepo(db.DB)
	votesRepo := repo.NewVoteRepo(db.DB)

	// 4) Servicios de dominio
	authSvc := auth.NewService(usersRepo, cfg.JWTSecret, cfg.JWTExpireMinutes)

	// 5) Productor de cola (SQS)
	var videoProducer queue.Producer
	{
		sqsURL := os.Getenv("SQS_QUEUE_URL")
		if sqsURL == "" {
			log.Fatal("SQS_QUEUE_URL no está definido (requerido para SQS)")
		}
		p, err := sqsqueue.NewProducer(context.Background(), sqsURL)
		if err != nil {
			log.Fatalf("Error creando SQS producer: %v", err)
		}
		videoProducer = p
		log.Printf("Usando SQS producer: %s", sqsURL)
	}
	defer videoProducer.Close()

	// 6) Redis (cache de rankings)
	redisCli := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	rankingsCache := cache.NewRankingsCache(redisCli, 3*time.Minute)

	// 7) Storage (S3 si hay bucket; local si no)
	var store storage.Storage
	if s3Bucket := os.Getenv("S3_BUCKET"); s3Bucket != "" {
		var err error
		store, err = storage.NewS3(s3Bucket)
		if err != nil {
			log.Fatalf("No se pudo inicializar S3 storage: %v", err)
		}
		log.Printf("Usando S3 bucket: %s", s3Bucket)
	} else {
		store = storage.NewLocal("./storage")
		log.Println("Usando storage local en ./storage")
	}

	// 8) Video service con productor de cola (SQS)
	videoSvc := videosvc.NewService(videosRepo, store, videoProducer)

	// 9) Handlers HTTP
	authH := httpapi.NewAuthHandlers(authSvc)
	videoH := httpapi.NewVideoHandlers(usersRepo, videosRepo, videoSvc)
	publicH := httpapi.NewPublicHandlers(videosRepo, votesRepo, usersRepo, rankingsCache)

	// 10) Router + CORS
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health
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
		api.POST("/public/videos/:id/vote", publicH.Vote) // votar requiere JWT
		api.DELETE("/auth", authH.DeleteUser)             // para tests
	}

	// Público sin auth
	r.GET("/api/public/videos", publicH.ListVideos)
	r.GET("/api/public/rankings", publicH.Rankings)
	r.GET("/api/public/cities", publicH.GetCities)

	log.Printf("API listening on :%s", cfg.AppPort)
	_ = r.Run(":" + cfg.AppPort)
}

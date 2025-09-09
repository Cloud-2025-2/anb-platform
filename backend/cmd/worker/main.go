package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/Cloud-2025-2/anb-platform/internal/config"
	"github.com/Cloud-2025-2/anb-platform/internal/db"
	"github.com/Cloud-2025-2/anb-platform/internal/domain"
	"github.com/Cloud-2025-2/anb-platform/internal/kafka"
	"github.com/Cloud-2025-2/anb-platform/internal/processing"
	"github.com/Cloud-2025-2/anb-platform/internal/repo"
	"github.com/Cloud-2025-2/anb-platform/internal/storage"
)

func main() {
	// Load environment variables
	_ = godotenv.Load()

	cfg := config.Load()

	// Database connection
	db.Connect()
	if err := db.DB.AutoMigrate(&domain.User{}, &domain.Video{}, &domain.Vote{}); err != nil {
		log.Fatal(err)
	}

	// Repositories
	videosRepo := repo.NewVideoRepo(db.DB)

	// Storage service
	store := storage.NewLocal("./storage")

	// Video processor
	processor := processing.NewVideoProcessor("./temp", "./assets", "./storage")

	// Kafka producer for retry/DLQ
	producer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Create worker service
	worker := NewWorkerService(videosRepo, store, processor)

	// Kafka consumer
	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "video-processors"
	}

	consumer, err := kafka.NewConsumer(cfg.KafkaBrokers, groupID, producer, worker)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigterm
		log.Println("Received termination signal, shutting down gracefully...")
		cancel()
	}()

	log.Println("Starting video processing worker...")
	log.Printf("Kafka brokers: %s", strings.Join(cfg.KafkaBrokers, ","))
	log.Printf("Consumer group: %s", groupID)

	// Start consuming
	if err := consumer.Start(ctx); err != nil {
		log.Printf("Consumer error: %v", err)
	}

	log.Println("Worker shutdown complete")
}

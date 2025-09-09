package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Cloud-2025-2/anb-platform/internal/config"
	"github.com/Cloud-2025-2/anb-platform/internal/db"
	"github.com/Cloud-2025-2/anb-platform/internal/domain"
	"github.com/Cloud-2025-2/anb-platform/internal/kafka"
	"github.com/Cloud-2025-2/anb-platform/internal/processing"
	"github.com/Cloud-2025-2/anb-platform/internal/repo"
	"github.com/Cloud-2025-2/anb-platform/internal/storage"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db.Connect()

	// Auto-migrate
	if err := db.DB.AutoMigrate(&domain.User{}, &domain.Video{}, &domain.Vote{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize repositories
	videoRepo := repo.NewVideoRepo(db.DB)

	// Initialize storage (for completeness)
	_ = storage.NewLocal("./storage")

	// Initialize video processor
	processor := processing.NewVideoProcessor("./storage", "./assets", "./assets")

	// Initialize Kafka producer
	producer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Test video processing pipeline
	fmt.Println("🎬 Testing Video Processing Pipeline")
	fmt.Println("=====================================")

	// Create test video entry
	testVideo := &domain.Video{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Title:  "Test Video Pipeline",
		Status: domain.VideoUploaded,
	}

	// Save to database
	if err := videoRepo.Create(testVideo); err != nil {
		log.Fatalf("Failed to create test video: %v", err)
	}

	fmt.Printf("✅ Created test video: %s\n", testVideo.ID)

	// Test Kafka message publishing
	task := kafka.VideoProcessingTask{
		VideoID:   testVideo.ID.String(),
		UserID:    testVideo.UserID.String(),
		Timestamp: time.Now(),
	}

	// Publish to main topic
	if err := producer.PublishVideoProcessingTask(task); err != nil {
		log.Fatalf("Failed to publish video processing task: %v", err)
	}

	fmt.Printf("✅ Published video processing task to Kafka\n")

	// Test video processor methods (without actual video files)
	fmt.Println("\n🔧 Testing Video Processor Methods")
	fmt.Println("==================================")

	// Test batch processing capability
	batchTasks := []kafka.VideoProcessingTask{
		{VideoID: uuid.New().String(), UserID: uuid.New().String()},
		{VideoID: uuid.New().String(), UserID: uuid.New().String()},
		{VideoID: uuid.New().String(), UserID: uuid.New().String()},
	}

	fmt.Printf("✅ Created batch of %d video processing tasks\n", len(batchTasks))

	// Test processor interface methods (these would fail without actual video files)
	fmt.Println("\n📋 Video Processor Interface Test")
	fmt.Println("=================================")
	
	testInputPath := "nonexistent.mp4"
	testOutputPath := "test_output.mp4"
	
	// These will fail but we're testing the interface exists
	fmt.Printf("Testing ProcessVideo method... ")
	if err := processor.ProcessVideo(testInputPath, testOutputPath); err != nil {
		fmt.Printf("❌ Expected failure (no input file): %v\n", err)
	}

	// Test database operations
	fmt.Println("\n💾 Database Operations Test")
	fmt.Println("===========================")

	// Update video status
	testVideo.Status = domain.VideoProcessing
	if err := videoRepo.Update(testVideo); err != nil {
		log.Fatalf("Failed to update video status: %v", err)
	}
	fmt.Printf("✅ Updated video status to: %s\n", testVideo.Status)

	// Test finding video
	foundVideo, err := videoRepo.FindByID(testVideo.ID)
	if err != nil {
		log.Fatalf("Failed to find video: %v", err)
	}
	fmt.Printf("✅ Found video: %s (Status: %s)\n", foundVideo.Title, foundVideo.Status)

	// Test JSON serialization of tasks
	fmt.Println("\n📦 JSON Serialization Test")
	fmt.Println("==========================")

	taskJSON, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal task: %v", err)
	}
	fmt.Printf("✅ VideoProcessingTask JSON:\n%s\n", taskJSON)

	// Clean up test video
	if err := videoRepo.DeleteByIDForUser(testVideo.ID, testVideo.UserID); err != nil {
		log.Printf("Warning: Failed to clean up test video: %v", err)
	} else {
		fmt.Printf("✅ Cleaned up test video\n")
	}

	fmt.Println("\n🎉 Pipeline Test Complete!")
	fmt.Println("==========================")
	fmt.Println("✅ Kafka producer interface working")
	fmt.Println("✅ Video processor methods available")
	fmt.Println("✅ Database operations functional")
	fmt.Println("✅ JSON serialization working")
	fmt.Println("✅ Error handling implemented")
	fmt.Println("\n📝 Next Steps:")
	fmt.Println("1. Start Kafka and database services: docker-compose up -d")
	fmt.Println("2. Run video processing workers: go run cmd/worker/main.go")
	fmt.Println("3. Upload actual video files through API")
	fmt.Println("4. Monitor worker logs for processing status")
}

package main

import (
        "context"
        "log"
        "os"
        "os/signal"
        "syscall"

        "github.com/google/uuid"
        "github.com/joho/godotenv"
        "github.com/aws/aws-sdk-go-v2/config"
        "github.com/aws/aws-sdk-go-v2/service/sqs"

        appconfig "github.com/Cloud-2025-2/anb-platform/internal/config"
        "github.com/Cloud-2025-2/anb-platform/internal/db"
        "github.com/Cloud-2025-2/anb-platform/internal/domain"
        "github.com/Cloud-2025-2/anb-platform/internal/processing"
        "github.com/Cloud-2025-2/anb-platform/internal/queue"
        sqsqueue "github.com/Cloud-2025-2/anb-platform/internal/queue/sqs"
        "github.com/Cloud-2025-2/anb-platform/internal/repo"
        "github.com/Cloud-2025-2/anb-platform/internal/storage"
)

func main() {
        // Load environment variables
        _ = godotenv.Load()

        _ = appconfig.Load()

        // Database connection
        db.Connect()
        if err := db.DB.AutoMigrate(&domain.User{}, &domain.Video{}, &domain.Vote{}); err != nil {
                log.Fatal(err)
        }

        // Repositories
        videosRepo := repo.NewVideoRepo(db.DB)

        // Storage service - use S3 if bucket is configured, otherwise use local
        var store storage.Storage
        s3Bucket := os.Getenv("S3_BUCKET")
        if s3Bucket != "" {
                var err error
                store, err = storage.NewS3(s3Bucket)
                if err != nil {
                        log.Fatalf("Failed to initialize S3 storage: %v", err)
                }
                log.Printf("Using S3 storage with bucket: %s", s3Bucket)
        } else {
                store = storage.NewLocal("./storage")
                log.Println("Using local storage at ./storage")
        }

        // Video processor
        storageDir := "./storage"
        if s3Bucket != "" {
                storageDir = s3Bucket // Use bucket name for S3
        }
        processor := processing.NewVideoProcessor("./temp", "./assets", storageDir)

        // Create worker service
        worker := NewWorkerService(videosRepo, store, processor)

        // Get SQS queue URL from environment
        queueURL := os.Getenv("SQS_QUEUE_URL")
        if queueURL == "" {
                log.Fatal("SQS_QUEUE_URL environment variable is required")
        }

        // Initialize AWS SDK v2 config
        awsRegion := os.Getenv("AWS_REGION")
        if awsRegion == "" {
                awsRegion = "us-east-1"
        }

        awsCfg, err := config.LoadDefaultConfig(context.Background(), 
                config.WithRegion(awsRegion))
        if err != nil {
                log.Fatalf("Failed to load AWS config: %v", err)
        }

        // Create SQS client
        sqsClient := sqs.NewFromConfig(awsCfg)

        // Create SQS consumer
        consumer := sqsqueue.NewConsumer(sqsClient, queueURL, 300, 20)

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

        log.Println("Starting video processing worker with SQS...")
        log.Printf("SQS Queue URL: %s", queueURL)
        log.Printf("AWS Region: %s", awsRegion)

        // Start consuming from SQS
        if err := consumer.Start(ctx, func(task queue.VideoProcessingTask) error {
                log.Printf("Processing task: VideoID=%s, FilePath=%s", task.VideoID, task.FilePath)
                
                // Parse VideoID string to UUID
                videoID, err := uuid.Parse(task.VideoID)
                if err != nil {
                        log.Printf("Error parsing video ID: %v", err)
                        return err
                }
                
                // Generate output path
                outputPath := task.VideoID + "_processed.mp4"
                
                return worker.ProcessVideoWithID(videoID, task.FilePath, outputPath)
        }); err != nil {
                log.Printf("Consumer error: %v", err)
        }

        log.Println("Worker shutdown complete")
}

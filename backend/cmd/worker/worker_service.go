package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/Cloud-2025-2/anb-platform/internal/domain"
	"github.com/Cloud-2025-2/anb-platform/internal/processing"
	"github.com/Cloud-2025-2/anb-platform/internal/repo"
	"github.com/Cloud-2025-2/anb-platform/internal/storage"
)

type WorkerService struct {
	videos    repo.VideoRepository
	store     storage.Storage
	processor *processing.VideoProcessor
}

func NewWorkerService(videos repo.VideoRepository, store storage.Storage, processor *processing.VideoProcessor) *WorkerService {
	return &WorkerService{
		videos:    videos,
		store:     store,
		processor: processor,
	}
}

// ProcessVideoWithID processes a video using the provided video ID
func (w *WorkerService) ProcessVideoWithID(videoID uuid.UUID, inputPath, outputPath string) error {
	log.Printf("Worker processing video: %s -> %s (VideoID: %s)", inputPath, outputPath, videoID)

	absInputPath, _ := filepath.Abs(inputPath)

	// Get video from database using the provided ID
	video, err := w.videos.FindByID(videoID)
	if err != nil {
		return fmt.Errorf("failed to find video in database: %w", err)
	}

	// Update status to processing
	video.Status = domain.VideoProcessing
	if err := w.videos.Update(video); err != nil {
		log.Printf("Warning: failed to update video status to processing: %v", err)
	}

	// Use the exact outputPath provided - don't generate a new one
	log.Printf("Processing video to output path: %s", outputPath)

	// Process the video using FFmpeg directly to the specified output path
	if err := w.processor.ProcessVideo(absInputPath, outputPath); err != nil {
		// Update status to failed
		video.Status = domain.VideoFailed
		w.videos.Update(video)
		return fmt.Errorf("video processing failed: %w", err)
	}

	// The processed file should already be at the correct location
	// Generate the URL based on the output path
	processedURL := fmt.Sprintf("/storage/%s", filepath.Base(outputPath))

	// Update video record with processed information
	now := time.Now()
	video.Status = domain.VideoPublished
	video.ProcessedURL = &processedURL
	video.ProcessedAt = &now
	video.PublishedAt = &now
	video.IsPublicForVote = true
	video.Watermark = true

	video.WidthProc = &[]int{1280}[0]
	video.HeightProc = &[]int{720}[0]
	video.AspectProc = &[]string{"16:9"}[0]
	video.HasAudioOrig = &[]bool{false}[0]

	if err := w.videos.Update(video); err != nil {
		return fmt.Errorf("failed to update video record: %w", err)
	}

	log.Printf("Successfully processed video %s to %s", video.ID, outputPath)
	return nil
}

// ProcessVideo processes a video by extracting ID from path (legacy method)
func (w *WorkerService) ProcessVideo(inputPath, outputPath string) error {
	log.Printf("Worker processing video: %s -> %s", inputPath, outputPath)

	// Extract video ID from the path (assuming it's in the filename)
	videoID, err := w.extractVideoIDFromPath(inputPath)
	if err != nil {
		return fmt.Errorf("failed to extract video ID: %w", err)
	}

	return w.ProcessVideoWithID(videoID, inputPath, outputPath)
}

func (w *WorkerService) extractVideoIDFromPath(path string) (uuid.UUID, error) {
	// Extract filename without extension
	filename := filepath.Base(path)
	nameWithoutExt := filename[:len(filename)-len(filepath.Ext(filename))]

	// Parse as UUID
	return uuid.Parse(nameWithoutExt)
}

func (w *WorkerService) generateProcessedFileName(videoID uuid.UUID) string {
	return fmt.Sprintf("%s_processed.mp4", videoID.String())
}

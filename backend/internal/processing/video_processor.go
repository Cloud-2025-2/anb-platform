package processing

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type VideoProcessor struct {
	tempDir    string
	assetsDir  string
	storageDir string
}

type ProcessingOptions struct {
	MaxDuration   int // seconds
	Width         int // 1280 for 720p
	Height        int // 720 for 720p
	WatermarkPath string
	IntroPath     string
	OutroPath     string
	RemoveAudio   bool
}

func NewVideoProcessor(tempDir, assetsDir, storageDir string) *VideoProcessor {
	return &VideoProcessor{
		tempDir:    tempDir,
		assetsDir:  assetsDir,
		storageDir: storageDir,
	}
}

func (vp *VideoProcessor) ProcessVideo(inputPath, outputPath string) error {
	log.Printf("Starting video processing: %s -> %s", inputPath, outputPath)

	// Check if input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	// Check if temp directory exists, create if not
	if err := os.MkdirAll(vp.tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create processing options
	opts := ProcessingOptions{
		MaxDuration:   30,
		Width:         1280,
		Height:        720,
		WatermarkPath: filepath.Join(vp.assetsDir, "logo.png"),
		IntroPath:     filepath.Join(vp.assetsDir, "intro.mp4"),
		OutroPath:     filepath.Join(vp.assetsDir, "outro.mp4"),
		RemoveAudio:   true,
	}

	// Step 1: Get video info and cut to max duration
	tempCut := filepath.Join(vp.tempDir, fmt.Sprintf("cut_%s.mp4", uuid.New().String()))
	if err := vp.cutVideo(inputPath, tempCut, opts.MaxDuration); err != nil {
		return fmt.Errorf("failed to cut video: %w", err)
	}
	defer os.Remove(tempCut)

	// Step 2: Adjust aspect ratio and resolution, remove audio
	tempResized := filepath.Join(vp.tempDir, fmt.Sprintf("resized_%s.mp4", uuid.New().String()))
	if err := vp.resizeAndRemoveAudio(tempCut, tempResized, opts); err != nil {
		return fmt.Errorf("failed to resize video: %w", err)
	}
	defer os.Remove(tempResized)

	// Step 3: Add watermark
	tempWatermarked := filepath.Join(vp.tempDir, fmt.Sprintf("watermarked_%s.mp4", uuid.New().String()))
	if err := vp.addWatermark(tempResized, tempWatermarked, opts.WatermarkPath); err != nil {
		return fmt.Errorf("failed to add watermark: %w", err)
	}
	defer os.Remove(tempWatermarked)

	// Step 4: Concatenate intro + main video + outro
	if err := vp.concatenateVideos(opts.IntroPath, tempWatermarked, opts.OutroPath, outputPath); err != nil {
		return fmt.Errorf("failed to concatenate videos: %w", err)
	}

	log.Printf("Video processing completed successfully: %s", outputPath)
	return nil
}

func (vp *VideoProcessor) cutVideo(inputPath, outputPath string, maxDuration int) error {
	log.Printf("Cutting video to %d seconds", maxDuration)

	return ffmpeg.Input(inputPath).
		Output(outputPath, ffmpeg.KwArgs{
			"t":                 strconv.Itoa(maxDuration),
			"c":                 "copy", // Copy streams without re-encoding when possible
			"avoid_negative_ts": "make_zero",
		}).
		OverWriteOutput().
		Run()
}

func (vp *VideoProcessor) resizeAndRemoveAudio(inputPath, outputPath string, opts ProcessingOptions) error {
	log.Printf("Resizing to %dx%d and removing audio", opts.Width, opts.Height)

	args := ffmpeg.KwArgs{
		"vf":     fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:black", opts.Width, opts.Height, opts.Width, opts.Height),
		"c:v":    "libx264",
		"crf":    "23",
		"preset": "medium",
		"r":      "30", // Force 30fps to maintain timing
		"an":     "",   // Remove audio
		"vsync":  "1",  // Ensure proper frame sync
	}

	return ffmpeg.Input(inputPath).
		Output(outputPath, args).
		OverWriteOutput().
		Run()
}

func (vp *VideoProcessor) addWatermark(inputPath, outputPath, watermarkPath string) error {
	log.Printf("Adding watermark: %s", watermarkPath)

	// Check if watermark exists
	if _, err := os.Stat(watermarkPath); os.IsNotExist(err) {
		return fmt.Errorf("watermark file not found: %s", watermarkPath)
	}

	// Position watermark at bottom-right corner with 10px padding
	return ffmpeg.Filter(
		[]*ffmpeg.Stream{
			ffmpeg.Input(inputPath),
			ffmpeg.Input(watermarkPath),
		}, "overlay", ffmpeg.Args{"main_w-overlay_w-10:main_h-overlay_h-10"}).
		Output(outputPath, ffmpeg.KwArgs{
			"c:v":    "libx264",
			"crf":    "23",
			"preset": "medium",
		}).
		OverWriteOutput().
		Run()
}

func (vp *VideoProcessor) concatenateVideos(introPath, mainPath, outroPath, outputPath string) error {
	log.Printf("Concatenating intro + main + outro videos")

	for _, path := range []string{introPath, mainPath, outroPath} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Printf("WARNING: File does not exist: %s", path)
			return fmt.Errorf("required file missing: %s", path)
		}
		log.Printf("File exists: %s", path)
	}

	absIntroPath, _ := filepath.Abs(introPath)
	absMainPath, _ := filepath.Abs(mainPath)
	absOutroPath, _ := filepath.Abs(outroPath)

	// Use filter_complex for better handling of different formats
	return ffmpeg.Filter(
		[]*ffmpeg.Stream{
			ffmpeg.Input(absIntroPath),
			ffmpeg.Input(absMainPath),
			ffmpeg.Input(absOutroPath),
		},
		"concat", ffmpeg.Args{"n=3:v=1:a=0"}, // 3 inputs, 1 video stream, 0 audio streams
	).Output(outputPath, ffmpeg.KwArgs{
		"c:v":    "libx264",
		"crf":    "23",
		"preset": "medium",
		"r":      "30",
		"an":     "", // No audio
	}).OverWriteOutput().Run()
}


func (vp *VideoProcessor) GetVideoInfo(inputPath string) (duration float64, width, height int, err error) {
	// This would typically use ffprobe, but for simplicity we'll return defaults
	// In a real implementation, you'd parse ffprobe output
	return 0, 0, 0, nil
}

func (vp *VideoProcessor) BatchProcess(tasks []string) error {
	log.Printf("Starting batch processing of %d videos", len(tasks))

	for i, _ := range tasks {
		log.Printf("Processing batch item %d/%d", i+1, len(tasks))
		// Process each video in the batch
		// Implementation would depend on your specific batch processing needs
	}

	return nil
}

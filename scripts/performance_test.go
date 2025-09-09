package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
	"time"
)

// Test configuration
const (
	API_BASE_URL     = "http://localhost:8000"
	CONCURRENT_USERS = 10
	VIDEOS_PER_USER  = 5
	TEST_VIDEO_PATH  = "test_video.mp4" // You need to provide this
)

type AuthResponse struct {
	Token string `json:"token"`
}

type VideoUploadResponse struct {
	VideoID string `json:"video_id"`
	Message string `json:"message"`
}

type TestResult struct {
	UserID       int
	VideoID      string
	UploadTime   time.Duration
	Success      bool
	ErrorMessage string
}

func main() {
	fmt.Println("🚀 ANB Platform Performance Test")
	fmt.Println("=================================")

	// Check if test video exists
	if _, err := os.Stat(TEST_VIDEO_PATH); os.IsNotExist(err) {
		log.Printf("⚠️  Test video file '%s' not found. Creating a dummy file for testing...", TEST_VIDEO_PATH)
		createDummyVideo()
	}

	// Create test users and get tokens
	tokens := make([]string, CONCURRENT_USERS)
	fmt.Printf("👥 Creating %d test users...\n", CONCURRENT_USERS)

	for i := 0; i < CONCURRENT_USERS; i++ {
		token, err := createTestUser(i)
		if err != nil {
			log.Fatalf("Failed to create user %d: %v", i, err)
		}
		tokens[i] = token
		fmt.Printf("✅ Created user %d\n", i)
	}

	// Performance test
	fmt.Printf("\n🎬 Starting video upload performance test...\n")
	fmt.Printf("Users: %d, Videos per user: %d, Total videos: %d\n", 
		CONCURRENT_USERS, VIDEOS_PER_USER, CONCURRENT_USERS*VIDEOS_PER_USER)

	results := make(chan TestResult, CONCURRENT_USERS*VIDEOS_PER_USER)
	var wg sync.WaitGroup

	startTime := time.Now()

	// Launch concurrent users
	for userID := 0; userID < CONCURRENT_USERS; userID++ {
		wg.Add(1)
		go func(uid int, token string) {
			defer wg.Done()
			
			for videoNum := 0; videoNum < VIDEOS_PER_USER; videoNum++ {
				uploadStart := time.Now()
				videoID, err := uploadVideo(token, uid, videoNum)
				uploadDuration := time.Since(uploadStart)

				result := TestResult{
					UserID:     uid,
					VideoID:    videoID,
					UploadTime: uploadDuration,
					Success:    err == nil,
				}

				if err != nil {
					result.ErrorMessage = err.Error()
				}

				results <- result
			}
		}(userID, tokens[userID])
	}

	// Wait for all uploads to complete
	wg.Wait()
	close(results)

	totalDuration := time.Since(startTime)

	// Analyze results
	var successful, failed int
	var totalUploadTime time.Duration
	var minTime, maxTime time.Duration
	minTime = time.Hour // Initialize with a large value

	fmt.Printf("\n📊 Test Results\n")
	fmt.Printf("===============\n")

	for result := range results {
		if result.Success {
			successful++
			totalUploadTime += result.UploadTime
			
			if result.UploadTime < minTime {
				minTime = result.UploadTime
			}
			if result.UploadTime > maxTime {
				maxTime = result.UploadTime
			}
			
			fmt.Printf("✅ User %d: Video uploaded in %v\n", result.UserID, result.UploadTime)
		} else {
			failed++
			fmt.Printf("❌ User %d: Upload failed - %s\n", result.UserID, result.ErrorMessage)
		}
	}

	// Performance metrics
	fmt.Printf("\n📈 Performance Metrics\n")
	fmt.Printf("======================\n")
	fmt.Printf("Total Duration: %v\n", totalDuration)
	fmt.Printf("Successful Uploads: %d\n", successful)
	fmt.Printf("Failed Uploads: %d\n", failed)
	fmt.Printf("Success Rate: %.2f%%\n", float64(successful)/float64(successful+failed)*100)

	if successful > 0 {
		avgUploadTime := totalUploadTime / time.Duration(successful)
		fmt.Printf("Average Upload Time: %v\n", avgUploadTime)
		fmt.Printf("Fastest Upload: %v\n", minTime)
		fmt.Printf("Slowest Upload: %v\n", maxTime)
		fmt.Printf("Throughput: %.2f uploads/second\n", float64(successful)/totalDuration.Seconds())
	}

	// Test Kafka message processing
	fmt.Printf("\n📨 Testing Kafka Message Processing\n")
	fmt.Printf("===================================\n")
	fmt.Printf("⏱️  Waiting 30 seconds for Kafka consumers to process messages...\n")
	
	time.Sleep(30 * time.Second)
	
	fmt.Printf("✅ Kafka processing test complete\n")
	fmt.Printf("💡 Check worker logs: docker-compose logs video-processor\n")

	// Cleanup
	fmt.Printf("\n🧹 Cleanup\n")
	fmt.Printf("==========\n")
	fmt.Printf("Note: Test users and videos remain in database for inspection\n")
	fmt.Printf("To clean up manually, restart the database container\n")

	fmt.Printf("\n🎉 Performance test completed!\n")
}

func createTestUser(userID int) (string, error) {
	userData := map[string]string{
		"email":    fmt.Sprintf("testuser%d@example.com", userID),
		"password": "testpassword123",
		"name":     fmt.Sprintf("Test User %d", userID),
	}

	jsonData, _ := json.Marshal(userData)

	// Register user
	resp, err := http.Post(API_BASE_URL+"/api/auth/signup", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	// Login to get token
	resp, err = http.Post(API_BASE_URL+"/api/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", err
	}

	return authResp.Token, nil
}

func uploadVideo(token string, userID, videoNum int) (string, error) {
	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add video file
	file, err := os.Open(TEST_VIDEO_PATH)
	if err != nil {
		return "", err
	}
	defer file.Close()

	part, err := writer.CreateFormFile("video_file", fmt.Sprintf("test_video_user%d_%d.mp4", userID, videoNum))
	if err != nil {
		return "", err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}

	// Add title
	writer.WriteField("title", fmt.Sprintf("Performance Test Video User%d #%d", userID, videoNum))

	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", API_BASE_URL+"/api/videos/upload", &buf)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	var uploadResp VideoUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return "", err
	}

	return uploadResp.VideoID, nil
}

func createDummyVideo() {
	// Create a minimal dummy video file for testing
	// This is just for testing purposes - in real scenarios you'd use actual video files
	dummyContent := []byte("This is a dummy video file for testing purposes")
	
	if err := os.WriteFile(TEST_VIDEO_PATH, dummyContent, 0644); err != nil {
		log.Fatalf("Failed to create dummy video file: %v", err)
	}
	
	fmt.Printf("✅ Created dummy video file: %s\n", TEST_VIDEO_PATH)
}

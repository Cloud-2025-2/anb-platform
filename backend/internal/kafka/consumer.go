package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"path/filepath"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

type Consumer struct {
	consumer      sarama.ConsumerGroup
	producer      *Producer
	processor     VideoProcessorInterface
	maxRetries    int
	baseBackoffMs int
}

type VideoProcessorInterface interface {
	ProcessVideo(inputPath, outputPath string) error
}

func NewConsumer(brokers []string, groupID string, producer *Producer, processor VideoProcessorInterface) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer:      consumer,
		producer:      producer,
		processor:     processor,
		maxRetries:    3,
		baseBackoffMs: 1000, // 1 second base backoff
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	topics := []string{TopicVideoProcessing, TopicVideoRetry}

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer context cancelled")
			return ctx.Err()
		default:
			if err := c.consumer.Consume(ctx, topics, c); err != nil {
				log.Printf("Error from consumer: %v", err)
				return err
			}
		}
	}
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	log.Println("Consumer group session setup")
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("Consumer group session cleanup")
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			log.Printf("Processing message from topic %s, partition %d, offset %d",
				message.Topic, message.Partition, message.Offset)

			if err := c.processMessage(message); err != nil {
				log.Printf("Error processing message: %v", err)
				// Don't mark as processed if there's an error
				continue
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func (c *Consumer) processMessage(message *sarama.ConsumerMessage) error {
	var task VideoProcessingTask
	if err := json.Unmarshal(message.Value, &task); err != nil {
		log.Printf("Failed to unmarshal task: %v", err)
		return err
	}

	log.Printf("Processing video task for VideoID: %s, RetryCount: %d", task.VideoID, task.RetryCount)

	// Apply exponential backoff for retry messages
	if message.Topic == TopicVideoRetry && task.RetryCount > 0 {
		backoffDuration := c.calculateBackoff(task.RetryCount)
		log.Printf("Applying backoff of %v for retry %d", backoffDuration, task.RetryCount)
		time.Sleep(backoffDuration)
	}

	// Process the video
	if err := c.processVideoTask(task); err != nil {
		return c.handleProcessingError(task, err)
	}

	log.Printf("Successfully processed video: %s", task.VideoID)
	return nil
}

func (c *Consumer) processVideoTask(task VideoProcessingTask) error {
	log.Printf("Processing video: %s", task.VideoID)

	// Parse video ID from string to UUID
	videoID, err := uuid.Parse(task.VideoID)
	if err != nil {
		return fmt.Errorf("invalid video ID format: %w", err)
	}

	// Generate proper output path using the original video ID
	inputDir := filepath.Dir(task.FilePath)
	outputPath := filepath.Join(inputDir, fmt.Sprintf("%s_processed.mp4", videoID.String()))

	// Call ProcessVideoWithID with the correct video ID from the Kafka message
	if processor, ok := c.processor.(interface {
		ProcessVideoWithID(uuid.UUID, string, string) error
	}); ok {
		return processor.ProcessVideoWithID(videoID, task.FilePath, outputPath)
	}

	// Fallback to old method if ProcessVideoWithID is not available
	return c.processor.ProcessVideo(task.FilePath, outputPath)
}

func (c *Consumer) handleProcessingError(task VideoProcessingTask, err error) error {
	log.Printf("Processing failed for video %s: %v", task.VideoID, err)

	if task.RetryCount >= c.maxRetries {
		log.Printf("Max retries exceeded for video %s, sending to DLQ", task.VideoID)
		return c.producer.PublishToDLQ(task, err.Error())
	}

	log.Printf("Retrying video %s (attempt %d/%d)", task.VideoID, task.RetryCount+1, c.maxRetries)
	return c.producer.PublishToRetryTopic(task)
}

func (c *Consumer) calculateBackoff(retryCount int) time.Duration {
	// Exponential backoff: baseBackoff * 2^retryCount with jitter
	backoffMs := c.baseBackoffMs * int(math.Pow(2, float64(retryCount)))

	// Add some jitter (±20%)
	jitter := int(float64(backoffMs) * 0.2)
	backoffMs += (retryCount*137)%(2*jitter) - jitter // Simple pseudo-random jitter

	return time.Duration(backoffMs) * time.Millisecond
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}

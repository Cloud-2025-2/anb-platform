package kafka

import (
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

type Producer struct {
	producer sarama.SyncProducer
}

type VideoProcessingTask struct {
	VideoID    string    `json:"video_id"`
	UserID     string    `json:"user_id"`
	Title      string    `json:"title"`
	FilePath   string    `json:"file_path"`
	Timestamp  time.Time `json:"timestamp"`
	RetryCount int       `json:"retry_count"`
}

const (
	TopicVideoProcessing = "video-processing"
	TopicVideoRetry      = "video-processing-retry"
	TopicVideoDLQ        = "video-processing-dlq"
)

func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{producer: producer}, nil
}

func (p *Producer) PublishVideoProcessingTask(task VideoProcessingTask) error {
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: TopicVideoProcessing,
		Key:   sarama.StringEncoder(task.VideoID),
		Value: sarama.ByteEncoder(taskBytes),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("task_id"),
				Value: []byte(uuid.New().String()),
			},
			{
				Key:   []byte("timestamp"),
				Value: []byte(task.Timestamp.Format(time.RFC3339)),
			},
		},
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Video processing task sent to partition %d at offset %d", partition, offset)
	return nil
}

func (p *Producer) PublishToRetryTopic(task VideoProcessingTask) error {
	task.RetryCount++
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: TopicVideoRetry,
		Key:   sarama.StringEncoder(task.VideoID),
		Value: sarama.ByteEncoder(taskBytes),
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}

func (p *Producer) PublishToDLQ(task VideoProcessingTask, errorMsg string) error {
	dlqTask := struct {
		VideoProcessingTask
		Error    string    `json:"error"`
		FailedAt time.Time `json:"failed_at"`
	}{
		VideoProcessingTask: task,
		Error:               errorMsg,
		FailedAt:            time.Now(),
	}

	taskBytes, err := json.Marshal(dlqTask)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: TopicVideoDLQ,
		Key:   sarama.StringEncoder(task.VideoID),
		Value: sarama.ByteEncoder(taskBytes),
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

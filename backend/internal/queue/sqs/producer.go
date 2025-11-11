package sqsqueue

import (
	"context"
	"encoding/json"

	"github.com/Cloud-2025-2/anb-platform/internal/queue"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Producer struct {
	c   *sqs.Client
	url string
}

func NewProducer(ctx context.Context, queueURL string) (*Producer, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &Producer{c: sqs.NewFromConfig(cfg), url: queueURL}, nil
}

func (p *Producer) PublishVideoProcessingTask(t queue.VideoProcessingTask) error {
	b, _ := json.Marshal(t)
	_, err := p.c.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    &p.url,
		MessageBody: aws.String(string(b)),
	})
	return err
}

// <-- faltaba esto
func (p *Producer) Close() error { return nil }

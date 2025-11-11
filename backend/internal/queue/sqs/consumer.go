package sqsqueue

import (
	"context"
	"encoding/json"

	"github.com/Cloud-2025-2/anb-platform/internal/queue"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Consumer struct {
	c               *sqs.Client
	url             string
	visSec, waitSec int32
}

func NewConsumer(c *sqs.Client, url string, vis, wait int32) *Consumer {
	return &Consumer{c: c, url: url, visSec: vis, waitSec: wait}
}

func (q *Consumer) Start(ctx context.Context, handle func(queue.VideoProcessingTask) error) error {
	for {
		out, err := q.c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl: &q.url, MaxNumberOfMessages: 10,
			WaitTimeSeconds: q.waitSec, VisibilityTimeout: q.visSec,
		})
		if err != nil {
			continue
		}
		for _, m := range out.Messages {
			var t queue.VideoProcessingTask
			_ = json.Unmarshal([]byte(*m.Body), &t)
			if err := handle(t); err == nil {
				_, _ = q.c.DeleteMessage(ctx, &sqs.DeleteMessageInput{
					QueueUrl: &q.url, ReceiptHandle: m.ReceiptHandle,
				})
			}
			// Si falla, dejamos que reintente y que el DLQ de SQS haga redrive
		}
	}
}
func (q *Consumer) Close() error { return nil }

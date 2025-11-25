output "queue_url" {
  description = "SQS queue URL"
  value       = aws_sqs_queue.video_processing.url
}

output "queue_arn" {
  description = "SQS queue ARN"
  value       = aws_sqs_queue.video_processing.arn
}

output "dlq_url" {
  description = "Dead Letter Queue URL"
  value       = aws_sqs_queue.video_processing_dlq.url
}

output "dlq_arn" {
  description = "Dead Letter Queue ARN"
  value       = aws_sqs_queue.video_processing_dlq.arn
}

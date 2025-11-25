resource "aws_sqs_queue" "video_processing" {
  name                       = "${var.project_name}-video-processing-queue"
  visibility_timeout_seconds = 300
  message_retention_seconds  = 86400
  max_message_size           = 262144
  delay_seconds              = 0
  receive_wait_time_seconds  = 20

  tags = {
    Name = "${var.project_name}-${var.environment}-video-processing-queue"
  }
}

resource "aws_sqs_queue" "video_processing_dlq" {
  name                       = "${var.project_name}-video-processing-dlq"
  message_retention_seconds  = 1209600 # 14 days

  tags = {
    Name = "${var.project_name}-${var.environment}-video-processing-dlq"
  }
}

resource "aws_sqs_queue_redrive_policy" "video_processing" {
  queue_url = aws_sqs_queue.video_processing.id

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.video_processing_dlq.arn
    maxReceiveCount     = 3
  })
}

output "backend_cpu_alarm_arn" {
  description = "Backend CPU alarm ARN"
  value       = aws_cloudwatch_metric_alarm.backend_cpu_high.arn
}

output "frontend_cpu_alarm_arn" {
  description = "Frontend CPU alarm ARN"
  value       = aws_cloudwatch_metric_alarm.frontend_cpu_high.arn
}

output "worker_cpu_alarm_arn" {
  description = "Worker CPU alarm ARN"
  value       = aws_cloudwatch_metric_alarm.worker_cpu_high.arn
}

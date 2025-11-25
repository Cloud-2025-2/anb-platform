# CloudWatch Alarms for Backend
resource "aws_cloudwatch_metric_alarm" "backend_cpu_high" {
  alarm_name          = "${var.project_name}-${var.environment}-backend-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/ECS"
  period              = 60
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "This metric monitors backend CPU utilization"
  alarm_actions       = []

  dimensions = {
    ClusterName = var.cluster_name
    ServiceName = var.backend_service
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-backend-cpu-high"
  }
}

# CloudWatch Alarms for Frontend
resource "aws_cloudwatch_metric_alarm" "frontend_cpu_high" {
  alarm_name          = "${var.project_name}-${var.environment}-frontend-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/ECS"
  period              = 60
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "This metric monitors frontend CPU utilization"
  alarm_actions       = []

  dimensions = {
    ClusterName = var.cluster_name
    ServiceName = var.frontend_service
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-frontend-cpu-high"
  }
}

# CloudWatch Alarms for Worker
resource "aws_cloudwatch_metric_alarm" "worker_cpu_high" {
  alarm_name          = "${var.project_name}-${var.environment}-worker-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/ECS"
  period              = 60
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "This metric monitors worker CPU utilization"
  alarm_actions       = []

  dimensions = {
    ClusterName = var.cluster_name
    ServiceName = var.worker_service
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-worker-cpu-high"
  }
}

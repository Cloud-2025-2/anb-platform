# ECS Cluster outputs
output "cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.main.name
}

output "cluster_id" {
  description = "ECS cluster ID"
  value       = aws_ecs_cluster.main.id
}

output "cluster_arn" {
  description = "ECS cluster ARN"
  value       = aws_ecs_cluster.main.arn
}

# Backend outputs
output "backend_task_definition_arn" {
  description = "Backend task definition ARN"
  value       = aws_ecs_task_definition.backend.arn
}

output "backend_service_name" {
  description = "Backend service name"
  value       = aws_ecs_service.backend.name
}

output "backend_service_id" {
  description = "Backend service ID"
  value       = aws_ecs_service.backend.id
}

# Frontend outputs
output "frontend_task_definition_arn" {
  description = "Frontend task definition ARN"
  value       = aws_ecs_task_definition.frontend.arn
}

output "frontend_service_name" {
  description = "Frontend service name"
  value       = aws_ecs_service.frontend.name
}

output "frontend_service_id" {
  description = "Frontend service ID"
  value       = aws_ecs_service.frontend.id
}

# Worker outputs
output "worker_task_definition_arn" {
  description = "Worker task definition ARN"
  value       = aws_ecs_task_definition.worker.arn
}

output "worker_service_name" {
  description = "Worker service name"
  value       = aws_ecs_service.worker.name
}

output "worker_service_id" {
  description = "Worker service ID"
  value       = aws_ecs_service.worker.id
}

# IAM Role outputs (reexpone las variables, NO recursos internos)
output "ecs_task_execution_role_arn" {
  description = "ECS task execution role ARN (existing role passed to module)"
  value       = var.ecs_task_execution_role_arn
}

output "ecs_task_role_arn" {
  description = "ECS task role ARN (existing role passed to module)"
  value       = var.ecs_task_role_arn
}

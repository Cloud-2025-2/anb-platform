output "vpc_id" {
  description = "VPC ID"
  value       = module.networking.vpc_id
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = module.networking.public_subnet_ids
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = module.networking.private_subnet_ids
}

output "frontend_alb_dns" {
  description = "Frontend ALB DNS name"
  value       = module.alb.frontend_alb_dns
}

output "backend_alb_dns" {
  description = "Backend ALB DNS name"
  value       = module.alb.backend_alb_dns
}

output "rds_endpoint" {
  description = "RDS endpoint"
  value       = module.rds.db_endpoint
  sensitive   = true
}

output "s3_bucket_name" {
  description = "S3 bucket name"
  value       = module.s3.bucket_name
}

output "sqs_queue_url" {
  description = "SQS queue URL"
  value       = module.sqs.queue_url
}

output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = module.ecs.cluster_name
}

output "backend_service_name" {
  description = "Backend service name"
  value       = module.ecs.backend_service_name
}

output "frontend_service_name" {
  description = "Frontend service name"
  value       = module.ecs.frontend_service_name
}

output "worker_service_name" {
  description = "Worker service name"
  value       = module.ecs.worker_service_name
}

# URLs para acceder a la aplicación
output "application_urls" {
  description = "Application access URLs"
  value = {
    frontend = "http://${module.alb.frontend_alb_dns}"
    backend  = "http://${module.alb.backend_alb_dns}"
    swagger  = "http://${module.alb.backend_alb_dns}/swagger/index.html"
  }
}

# Información de conexión a base de datos
output "database_connection" {
  description = "Database connection information"
  value = {
    endpoint = module.rds.db_endpoint
    port     = module.rds.db_port
    name     = var.db_name
  }
  sensitive = true
}

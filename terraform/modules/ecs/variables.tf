variable "project_name" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "private_subnet_ids" {
  description = "Private subnet IDs"
  type        = list(string)
}

variable "ecs_security_group_id" {
  description = "ECS security group ID"
  type        = string
}

# ECR Images
variable "backend_image_uri" {
  description = "Backend Docker image URI"
  type        = string
}

variable "frontend_image_uri" {
  description = "Frontend Docker image URI"
  type        = string
}

variable "worker_image_uri" {
  description = "Worker Docker image URI"
  type        = string
}

# Load Balancers
variable "frontend_target_group_arn" {
  description = "Frontend target group ARN"
  type        = string
}

variable "backend_target_group_arn" {
  description = "Backend target group ARN"
  type        = string
}

# Database
variable "db_host" {
  description = "Database host"
  type        = string
}

variable "db_name" {
  description = "Database name"
  type        = string
}

variable "db_username" {
  description = "Database username"
  type        = string
  sensitive   = true
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "db_port" {
  description = "Database port"
  type        = number
}

# S3 and SQS
variable "s3_bucket_name" {
  description = "S3 bucket name"
  type        = string
}

variable "sqs_queue_url" {
  description = "SQS queue URL"
  type        = string
}

# Backend ALB DNS for frontend proxy
variable "backend_alb_dns" {
  description = "Backend ALB DNS name"
  type        = string
}

# Auto Scaling
variable "backend_min_capacity" {
  description = "Minimum number of backend tasks"
  type        = number
  default     = 1
}

variable "backend_max_capacity" {
  description = "Maximum number of backend tasks"
  type        = number
  default     = 3
}

variable "frontend_min_capacity" {
  description = "Minimum number of frontend tasks"
  type        = number
  default     = 1
}

variable "frontend_max_capacity" {
  description = "Maximum number of frontend tasks"
  type        = number
  default     = 3
}

variable "worker_min_capacity" {
  description = "Minimum number of worker tasks"
  type        = number
  default     = 1
}

variable "worker_max_capacity" {
  description = "Maximum number of worker tasks"
  type        = number
  default     = 3
}

variable "cpu_target_value" {
  description = "Target CPU utilization for auto scaling"
  type        = number
  default     = 70
}

# 👉 NUEVO: ARNs de roles IAM EXISTENTES (NO se crean en Terraform)
variable "ecs_task_execution_role_arn" {
  description = "ARN of existing IAM role for ECS task execution"
  type        = string
}

variable "ecs_task_role_arn" {
  description = "ARN of existing IAM role for ECS tasks (app permissions)"
  type        = string
}

variable "aws_region" {
  description = "AWS Region"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name"
  type        = string
  default     = "anb-platform"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.0.0.0/16"
}

# Database variables
variable "db_username" {
  description = "Database master username"
  type        = string
  default     = "postgres"
  sensitive   = true
}

variable "db_password" {
  description = "Database master password"
  type        = string
  sensitive   = true
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "anb_platform"
}

# ECR Image URIs
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

# Auto Scaling configuration
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

# 🔥 NUEVO: ARNs de los roles IAM EXISTENTES para ECS (NO se crean en Terraform)
variable "ecs_task_execution_role_arn" {
  description = "ARN of existing IAM role for ECS task execution"
  type        = string
}

variable "ecs_task_role_arn" {
  description = "ARN of existing IAM role for ECS tasks (application permissions)"
  type        = string
}

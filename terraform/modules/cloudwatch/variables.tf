variable "project_name" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "cluster_name" {
  description = "ECS cluster name"
  type        = string
}

variable "backend_service" {
  description = "Backend service name"
  type        = string
}

variable "frontend_service" {
  description = "Frontend service name"
  type        = string
}

variable "worker_service" {
  description = "Worker service name"
  type        = string
}

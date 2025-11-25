terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Backend para estado remoto (opcional)
  # backend "s3" {
  #   bucket = "anb-terraform-state"
  #   key    = "anb-platform/terraform.tfstate"
  #   region = "us-east-1"
  # }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}

# Data sources
data "aws_caller_identity" "current" {}
data "aws_availability_zones" "available" {
  state = "available"
}

# Variables locales
locals {
  account_id = data.aws_caller_identity.current.account_id
  azs        = slice(data.aws_availability_zones.available.names, 0, 2)
  
  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# Networking Module
module "networking" {
  source = "./modules/networking"

  project_name = var.project_name
  environment  = var.environment
  vpc_cidr     = var.vpc_cidr
  azs          = local.azs
}

# Security Groups Module
module "security" {
  source = "./modules/security"

  project_name = var.project_name
  environment  = var.environment
  vpc_id       = module.networking.vpc_id
}

# S3 Module
module "s3" {
  source = "./modules/s3"

  project_name = var.project_name
  environment  = var.environment
  account_id   = local.account_id
}

# SQS Module
module "sqs" {
  source = "./modules/sqs"

  project_name = var.project_name
  environment  = var.environment
}

# RDS Module
module "rds" {
  source = "./modules/rds"

  project_name        = var.project_name
  environment         = var.environment
  vpc_id              = module.networking.vpc_id
  private_subnet_ids  = module.networking.private_subnet_ids
  db_security_group_id = module.security.rds_security_group_id
  db_username         = var.db_username
  db_password         = var.db_password
  db_name             = var.db_name
}

# Application Load Balancers Module
module "alb" {
  source = "./modules/alb"

  project_name        = var.project_name
  environment         = var.environment
  vpc_id              = module.networking.vpc_id
  public_subnet_ids   = module.networking.public_subnet_ids
  alb_security_group_id = module.security.alb_security_group_id
}

# ECS Cluster and Services Module
module "ecs" {
  source = "./modules/ecs"

  project_name           = var.project_name
  environment            = var.environment
  vpc_id                 = module.networking.vpc_id
  private_subnet_ids     = module.networking.private_subnet_ids
  ecs_security_group_id  = module.security.ecs_security_group_id
  
  # ECR repositories
  backend_image_uri      = var.backend_image_uri
  frontend_image_uri     = var.frontend_image_uri
  worker_image_uri       = var.worker_image_uri
  
  # Load Balancers
  frontend_target_group_arn = module.alb.frontend_target_group_arn
  backend_target_group_arn  = module.alb.backend_target_group_arn
  
  # Database
  db_host                = module.rds.db_endpoint
  db_name                = var.db_name
  db_username            = var.db_username
  db_password            = var.db_password
  db_port                = module.rds.db_port
  
  # S3 and SQS
  s3_bucket_name         = module.s3.bucket_name
  sqs_queue_url          = module.sqs.queue_url
  
  # Backend ALB DNS for frontend proxy
  backend_alb_dns        = module.alb.backend_alb_dns
}

# CloudWatch Module
module "cloudwatch" {
  source = "./modules/cloudwatch"

  project_name     = var.project_name
  environment      = var.environment
  cluster_name     = module.ecs.cluster_name
  backend_service  = module.ecs.backend_service_name
  frontend_service = module.ecs.frontend_service_name
  worker_service   = module.ecs.worker_service_name
}

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

  # 👉 NUEVO: pasamos los ARNs de roles IAM EXISTENTES (no los crea Terraform)
  ecs_task_execution_role_arn = var.ecs_task_execution_role_arn
  ecs_task_role_arn           = var.ecs_task_role_arn
}

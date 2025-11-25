resource "aws_db_subnet_group" "main" {
  name       = "${var.project_name}-${var.environment}-db-subnet-group"
  subnet_ids = var.private_subnet_ids

  tags = {
    Name = "${var.project_name}-${var.environment}-db-subnet-group"
  }
}

resource "aws_db_instance" "postgres" {
  identifier        = "${var.project_name}-${var.environment}-postgres"
  engine            = "postgres"
  engine_version    = "15.4"
  instance_class    = "db.t3.micro"
  allocated_storage = 20
  storage_type      = "gp2"

  db_name  = var.db_name
  username = var.db_username
  password = var.db_password
  port     = 5432

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [var.db_security_group_id]
  publicly_accessible    = false

  # Multi-AZ for high availability
  multi_az = false # Set to true for production

  # Backup configuration
  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "mon:04:00-mon:05:00"

  # Snapshot configuration
  skip_final_snapshot       = true
  final_snapshot_identifier = "${var.project_name}-${var.environment}-final-snapshot"

  # Performance Insights (optional)
  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]

  tags = {
    Name = "${var.project_name}-${var.environment}-postgres"
  }
}

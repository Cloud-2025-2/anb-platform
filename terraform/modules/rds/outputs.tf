output "db_endpoint" {
  description = "RDS endpoint"
  value       = aws_db_instance.postgres.endpoint
}

output "db_address" {
  description = "RDS address"
  value       = aws_db_instance.postgres.address
}

output "db_port" {
  description = "RDS port"
  value       = aws_db_instance.postgres.port
}

output "db_name" {
  description = "Database name"
  value       = aws_db_instance.postgres.db_name
}

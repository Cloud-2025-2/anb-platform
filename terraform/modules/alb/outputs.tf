output "frontend_alb_arn" {
  description = "Frontend ALB ARN"
  value       = aws_lb.frontend.arn
}

output "frontend_alb_dns" {
  description = "Frontend ALB DNS name"
  value       = aws_lb.frontend.dns_name
}

output "frontend_target_group_arn" {
  description = "Frontend target group ARN"
  value       = aws_lb_target_group.frontend.arn
}

output "backend_alb_arn" {
  description = "Backend ALB ARN"
  value       = aws_lb.backend.arn
}

output "backend_alb_dns" {
  description = "Backend ALB DNS name"
  value       = aws_lb.backend.dns_name
}

output "backend_target_group_arn" {
  description = "Backend target group ARN"
  value       = aws_lb_target_group.backend.arn
}

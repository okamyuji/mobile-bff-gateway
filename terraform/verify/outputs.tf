output "alb_dns_name" {
  value       = var.deploy_workload ? aws_lb.this[0].dns_name : null
  description = "検証用ALBのDNS名 (HTTP only)"
}

output "dns_name" {
  value       = aws_lb.this.dns_name
  description = "ALBのDNS名です。"
}

output "target_group_arn" {
  value       = aws_lb_target_group.gateway.arn
  description = "Gateway用ターゲットグループARNです。"
}

output "security_group_id" {
  value       = aws_security_group.this.id
  description = "ALBのセキュリティグループIDです。"
}

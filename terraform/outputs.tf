output "alb_dns_name" {
  value       = module.alb.dns_name
  description = "ALBのDNS名です。"
}

output "cloudfront_domain_name" {
  value       = module.edge.domain_name
  description = "CloudFrontのDNS名です。"
}

output "domain_name" {
  value       = aws_cloudfront_distribution.this.domain_name
  description = "CloudFrontのDNS名です。"
}

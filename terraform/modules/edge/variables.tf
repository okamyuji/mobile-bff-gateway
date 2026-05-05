variable "name" {
  type        = string
  description = "リソース名の接頭辞を指定します。"
}

variable "alb_domain_name" {
  type        = string
  description = "CloudFrontのオリジンにするALBのDNS名を指定します。"
}

variable "web_acl_name" {
  type        = string
  description = "CloudFront用WAF Web ACL名を指定します。"
}

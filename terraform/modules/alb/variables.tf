variable "name" {
  type        = string
  description = "リソース名の接頭辞を指定します。"
}

variable "vpc_id" {
  type        = string
  description = "VPC IDを指定します。"
}

variable "public_subnets" {
  type        = list(string)
  description = "ALB用のパブリックサブネットIDを指定します。"
}

variable "certificate_arn" {
  type        = string
  description = "HTTPS終端に使うACM証明書ARNを指定します。"
}

variable "allowed_cidrs" {
  type        = list(string)
  description = "ALBへのHTTPS接続を許可するCIDRを指定します。"
}

variable "aws_region" {
  type    = string
  default = "ap-northeast-1"
}

variable "aws_profile" {
  type    = string
  default = "fintech-apigw"
}

variable "name_prefix" {
  type    = string
  default = "mbg-verify"
}

variable "container_image" {
  type        = string
  description = "ECRにpush済みのGateway/Mock共通イメージURI (例: <acct>.dkr.ecr.ap-northeast-1.amazonaws.com/mbg-verify:latest)"
  default     = ""
}

variable "deploy_workload" {
  type        = bool
  description = "true でALB+ECSサービスをデプロイ。falseならECRのみ作成 (画像push前の段階で使用)"
  default     = false
}

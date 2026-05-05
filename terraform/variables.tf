variable "aws_region" {
  type        = string
  description = "AWSリージョンを指定します。"
  default     = "ap-northeast-1"
}

variable "aws_profile" {
  type        = string
  description = "Terraform実行に使うAWS CLIプロファイル名を指定します。"
  default     = "fintech-apigw"
}

variable "name" {
  type        = string
  description = "リソース名の接頭辞を指定します。"
  default     = "mobile-bff-gateway"
}

variable "vpc_id" {
  type        = string
  description = "既存VPC IDを指定します。"
}

variable "public_subnets" {
  type        = list(string)
  description = "ALBを配置する2つ以上のパブリックサブネットIDを指定します。"
}

variable "private_subnets" {
  type        = list(string)
  description = "ECSタスクを配置する2つ以上のプライベートサブネットIDを指定します。"
}

variable "certificate_arn" {
  type        = string
  description = "ALBのHTTPSリスナーで使うACM証明書ARNを指定します。"
}

variable "allowed_cidrs" {
  type        = list(string)
  description = "ALBへ到達できるCIDRを指定します。CloudFrontのみに絞る場合は管理プレフィックスリストを追加してください。"
  default     = ["0.0.0.0/0"]
}

variable "container_image" {
  type        = string
  description = "Gatewayコンテナイメージを指定します。"
}

variable "desired_count" {
  type        = number
  description = "ECSサービスの希望タスク数を指定します。"
  default     = 2
}

variable "cpu" {
  type        = number
  description = "FargateタスクCPUを指定します。"
  default     = 512
}

variable "memory" {
  type        = number
  description = "Fargateタスクメモリを指定します。"
  default     = 1024
}

variable "user_service_url" {
  type        = string
  description = "ユーザーサービスの内部URLを指定します。"
}

variable "payment_service_url" {
  type        = string
  description = "決済注文サービスの内部URLを指定します。"
}

variable "account_service_url" {
  type        = string
  description = "口座サービスの内部URLを指定します。"
}

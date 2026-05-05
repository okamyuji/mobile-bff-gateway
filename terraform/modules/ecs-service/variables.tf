variable "name" {
  type        = string
  description = "リソース名の接頭辞を指定します。"
}

variable "vpc_id" {
  type        = string
  description = "VPC IDを指定します。"
}

variable "private_subnets" {
  type        = list(string)
  description = "Fargateタスク用のプライベートサブネットIDを指定します。"
}

variable "target_group_arn" {
  type        = string
  description = "ALBターゲットグループARNを指定します。"
}

variable "container_image" {
  type        = string
  description = "Gatewayコンテナイメージを指定します。"
}

variable "desired_count" {
  type        = number
  description = "希望タスク数を指定します。"
}

variable "cpu" {
  type        = number
  description = "FargateタスクCPUを指定します。"
}

variable "memory" {
  type        = number
  description = "Fargateタスクメモリを指定します。"
}

variable "alb_security_group_id" {
  type        = string
  description = "ALBのセキュリティグループIDを指定します。"
}

variable "environment" {
  type        = map(string)
  description = "Gatewayコンテナへ渡す環境変数を指定します。"
}

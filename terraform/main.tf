terraform {
  required_version = ">= 1.8.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region  = var.aws_region
  profile = var.aws_profile
}

provider "aws" {
  alias   = "use1"
  region  = "us-east-1"
  profile = var.aws_profile
}

module "alb" {
  source = "./modules/alb"

  name            = var.name
  vpc_id          = var.vpc_id
  public_subnets  = var.public_subnets
  certificate_arn = var.certificate_arn
  allowed_cidrs   = var.allowed_cidrs
}

module "ecs_gateway" {
  source = "./modules/ecs-service"

  name                  = var.name
  vpc_id                = var.vpc_id
  private_subnets       = var.private_subnets
  target_group_arn      = module.alb.target_group_arn
  container_image       = var.container_image
  desired_count         = var.desired_count
  cpu                   = var.cpu
  memory                = var.memory
  alb_security_group_id = module.alb.security_group_id

  environment = {
    USER_SERVICE_URL    = var.user_service_url
    PAYMENT_SERVICE_URL = var.payment_service_url
    ACCOUNT_SERVICE_URL = var.account_service_url
    DOWNSTREAM_TIMEOUT  = "800ms"
    CACHE_TTL           = "5s"
    RATE_LIMIT_BURST    = "100"
    RATE_LIMIT_REFILL   = "100"
    RATE_LIMIT_EVERY    = "1s"
  }
}

module "edge" {
  source = "./modules/edge"

  providers = {
    aws = aws.use1
  }

  name            = var.name
  alb_domain_name = module.alb.dns_name
  web_acl_name    = "${var.name}-web-acl"
}

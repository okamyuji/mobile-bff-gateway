resource "aws_cloudwatch_log_group" "this" {
  count             = var.deploy_workload ? 1 : 0
  name              = "/ecs/${var.name_prefix}"
  retention_in_days = 7
}

resource "aws_ecs_cluster" "this" {
  count = var.deploy_workload ? 1 : 0
  name  = var.name_prefix
}

resource "aws_security_group" "task" {
  count       = var.deploy_workload ? 1 : 0
  name        = "${var.name_prefix}-task"
  description = "ECS task SG for verification (gateway+mocks)"
  vpc_id      = aws_vpc.this[0].id

  ingress {
    description     = "From ALB"
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb[0].id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.name_prefix}-task"
  }
}

resource "aws_iam_role" "execution" {
  count = var.deploy_workload ? 1 : 0
  name  = "${var.name_prefix}-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "ecs-tasks.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "execution" {
  count      = var.deploy_workload ? 1 : 0
  role       = aws_iam_role.execution[0].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

data "aws_region" "current" {}

locals {
  log_group_name = var.deploy_workload ? aws_cloudwatch_log_group.this[0].name : ""

  mocks = [
    { name = "user-service", kind = "user", port = 8081 },
    { name = "payment-service", kind = "payment", port = 8082 },
    { name = "account-service", kind = "account", port = 8083 },
  ]

  gateway_container = {
    name      = "gateway"
    image     = var.container_image
    essential = true
    command   = ["/usr/local/bin/gateway"]
    portMappings = [{
      containerPort = 8080
      protocol      = "tcp"
    }]
    environment = [
      { name = "ADDR", value = ":8080" },
      { name = "USER_SERVICE_URL", value = "http://localhost:8081/user" },
      { name = "PAYMENT_SERVICE_URL", value = "http://localhost:8082/payment-orders" },
      { name = "ACCOUNT_SERVICE_URL", value = "http://localhost:8083/account" },
      { name = "DOWNSTREAM_TIMEOUT", value = "1500ms" },
      { name = "CACHE_TTL", value = "5s" },
      { name = "RATE_LIMIT_BURST", value = "100" },
      { name = "RATE_LIMIT_REFILL", value = "100" },
      { name = "RATE_LIMIT_EVERY", value = "1s" },
    ]
    dependsOn = [
      { containerName = "user-service", condition = "START" },
      { containerName = "payment-service", condition = "START" },
      { containerName = "account-service", condition = "START" },
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        awslogs-group         = local.log_group_name
        awslogs-region        = data.aws_region.current.name
        awslogs-stream-prefix = "gateway"
      }
    }
  }

  mock_containers = [
    for m in local.mocks : {
      name       = m.name
      image      = var.container_image
      essential  = true
      entryPoint = ["/usr/local/bin/mockservice"]
      portMappings = [{
        containerPort = m.port
        protocol      = "tcp"
      }]
      environment = [
        { name = "SERVICE_KIND", value = m.kind },
        { name = "ADDR", value = ":${m.port}" },
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = local.log_group_name
          awslogs-region        = data.aws_region.current.name
          awslogs-stream-prefix = m.name
        }
      }
    }
  ]

  containers = concat([local.gateway_container], local.mock_containers)
}

resource "aws_ecs_task_definition" "this" {
  count                    = var.deploy_workload ? 1 : 0
  family                   = var.name_prefix
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 1024
  memory                   = 2048
  execution_role_arn       = aws_iam_role.execution[0].arn

  runtime_platform {
    cpu_architecture        = "X86_64"
    operating_system_family = "LINUX"
  }

  container_definitions = jsonencode(local.containers)
}

resource "aws_ecs_service" "this" {
  count           = var.deploy_workload ? 1 : 0
  name            = var.name_prefix
  cluster         = aws_ecs_cluster.this[0].id
  task_definition = aws_ecs_task_definition.this[0].arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets = [
      aws_subnet.public_a[0].id,
      aws_subnet.public_c[0].id,
    ]
    security_groups  = [aws_security_group.task[0].id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.gateway[0].arn
    container_name   = "gateway"
    container_port   = 8080
  }

  depends_on = [aws_lb_listener.http]
}

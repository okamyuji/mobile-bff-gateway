resource "aws_security_group" "alb" {
  count       = var.deploy_workload ? 1 : 0
  name        = "${var.name_prefix}-alb"
  description = "ALB security group for mobile-bff-gateway verification"
  vpc_id      = aws_vpc.this[0].id

  ingress {
    description = "HTTP from anywhere (verification only)"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.name_prefix}-alb"
  }
}

resource "aws_lb" "this" {
  count              = var.deploy_workload ? 1 : 0
  name               = "${var.name_prefix}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb[0].id]
  subnets = [
    aws_subnet.public_a[0].id,
    aws_subnet.public_c[0].id,
  ]

  tags = {
    Name = "${var.name_prefix}-alb"
  }
}

resource "aws_lb_target_group" "gateway" {
  count       = var.deploy_workload ? 1 : 0
  name        = "${var.name_prefix}-tg"
  port        = 8080
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = aws_vpc.this[0].id

  health_check {
    enabled             = true
    path                = "/healthz"
    matcher             = "200"
    interval            = 15
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 2
  }
}

resource "aws_lb_listener" "http" {
  count             = var.deploy_workload ? 1 : 0
  load_balancer_arn = aws_lb.this[0].arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.gateway[0].arn
  }
}

resource "aws_vpc" "this" {
  count                = var.deploy_workload ? 1 : 0
  cidr_block           = "10.99.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${var.name_prefix}-vpc"
  }
}

resource "aws_internet_gateway" "this" {
  count  = var.deploy_workload ? 1 : 0
  vpc_id = aws_vpc.this[0].id

  tags = {
    Name = "${var.name_prefix}-igw"
  }
}

resource "aws_subnet" "public_a" {
  count                   = var.deploy_workload ? 1 : 0
  vpc_id                  = aws_vpc.this[0].id
  cidr_block              = "10.99.0.0/24"
  availability_zone       = "ap-northeast-1a"
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.name_prefix}-public-a"
  }
}

resource "aws_subnet" "public_c" {
  count                   = var.deploy_workload ? 1 : 0
  vpc_id                  = aws_vpc.this[0].id
  cidr_block              = "10.99.1.0/24"
  availability_zone       = "ap-northeast-1c"
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.name_prefix}-public-c"
  }
}

resource "aws_route_table" "public" {
  count  = var.deploy_workload ? 1 : 0
  vpc_id = aws_vpc.this[0].id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.this[0].id
  }

  tags = {
    Name = "${var.name_prefix}-public-rt"
  }
}

resource "aws_route_table_association" "public_a" {
  count          = var.deploy_workload ? 1 : 0
  subnet_id      = aws_subnet.public_a[0].id
  route_table_id = aws_route_table.public[0].id
}

resource "aws_route_table_association" "public_c" {
  count          = var.deploy_workload ? 1 : 0
  subnet_id      = aws_subnet.public_c[0].id
  route_table_id = aws_route_table.public[0].id
}

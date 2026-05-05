resource "aws_ecr_repository" "this" {
  name                 = var.name_prefix
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = false
  }
}

output "ecr_repository_url" {
  value = aws_ecr_repository.this.repository_url
}

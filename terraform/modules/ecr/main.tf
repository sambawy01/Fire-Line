variable "environment" { type = string }

resource "aws_ecr_repository" "fireline" {
  name                 = "fireline-${var.environment}"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }
}

resource "aws_ecr_lifecycle_policy" "fireline" {
  repository = aws_ecr_repository.fireline.name

  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "Keep last 20 images"
      selection = {
        tagStatus   = "any"
        countType   = "imageCountMoreThan"
        countNumber = 20
      }
      action = { type = "expire" }
    }]
  })
}

output "repository_url" { value = aws_ecr_repository.fireline.repository_url }
output "repository_arn" { value = aws_ecr_repository.fireline.arn }

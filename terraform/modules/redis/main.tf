variable "environment" { type = string }
variable "vpc_id" { type = string }
variable "private_subnet_ids" { type = list(string) }
variable "node_type" { type = string }
variable "ecs_security_group_id" { type = string }

resource "aws_elasticache_subnet_group" "main" {
  name       = "fireline-${var.environment}"
  subnet_ids = var.private_subnet_ids
}

resource "aws_security_group" "redis" {
  name_prefix = "fireline-${var.environment}-redis-"
  vpc_id      = var.vpc_id

  ingress {
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [var.ecs_security_group_id]
    description     = "Redis from ECS"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "fireline-${var.environment}-redis-sg" }
}

resource "aws_elasticache_replication_group" "main" {
  replication_group_id = "fireline-${var.environment}"
  description          = "FireLine Redis - ${var.environment}"
  node_type            = var.node_type
  num_cache_clusters   = var.environment == "production" ? 2 : 1
  engine_version       = "7.0"
  port                 = 6379

  subnet_group_name  = aws_elasticache_subnet_group.main.name
  security_group_ids = [aws_security_group.redis.id]

  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  automatic_failover_enabled = var.environment == "production"

  snapshot_retention_limit = var.environment == "production" ? 7 : 1
  snapshot_window          = "03:00-04:00"
  maintenance_window       = "sun:04:00-sun:05:00"

  tags = { Name = "fireline-${var.environment}" }
}

output "endpoint" {
  value = aws_elasticache_replication_group.main.primary_endpoint_address
}

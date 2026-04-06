variable "environment" { type = string }

resource "aws_secretsmanager_secret" "db_credentials" {
  name        = "fireline/${var.environment}/db-credentials"
  description = "FireLine database credentials"
  tags        = { Name = "fireline-${var.environment}-db-creds" }
}

resource "aws_secretsmanager_secret" "jwt_key" {
  name        = "fireline/${var.environment}/jwt-private-key"
  description = "FireLine JWT RSA private key"
  tags        = { Name = "fireline-${var.environment}-jwt-key" }
}

resource "aws_secretsmanager_secret" "admin_db_url" {
  name        = "fireline/${var.environment}/admin-database-url"
  description = "FireLine admin (superuser) database URL"
  tags        = { Name = "fireline-${var.environment}-admin-db-url" }
}

resource "aws_secretsmanager_secret" "adapter_secrets" {
  name        = "fireline/${var.environment}/adapter-secrets"
  description = "FireLine POS adapter credentials"
  tags        = { Name = "fireline-${var.environment}-adapter-secrets" }
}

output "db_secret_arn" { value = aws_secretsmanager_secret.db_credentials.arn }
output "admin_db_secret_arn" { value = aws_secretsmanager_secret.admin_db_url.arn }
output "jwt_secret_arn" { value = aws_secretsmanager_secret.jwt_key.arn }
output "adapter_secret_arn" { value = aws_secretsmanager_secret.adapter_secrets.arn }

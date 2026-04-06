variable "database_url" {
  type    = string
  default = "postgres://fireline:fireline@localhost:5432/fireline?sslmode=disable"
}

env "local" {
  src = "file://migrations"
  url = var.database_url
  dev = "docker://postgres/16/dev"
}

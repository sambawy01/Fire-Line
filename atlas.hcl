env "local" {
  src = "file://migrations"
  url = "postgres://fireline:fireline@localhost:5432/fireline?sslmode=disable"
  dev = "docker://postgres/16/dev"
}

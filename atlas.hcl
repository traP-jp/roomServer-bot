env "local" {
  src = "file://internal/infrastructure/mariadb/schema/schema.ma.hcl"
  url = "maria://user:password@localhost:3306/database"
  dev = "docker://maria/latest/dev"
}

env "local" {
  src = "file://internal/infrastructure/mariadb/schema/schema.sql"
  url = "maria://user:password@localhost:3306/database"
  dev = "maria://user:password@localhost:3306/dev"
  format {
    migrate {
      diff = "{{ sql . }}"
    }
  }
}

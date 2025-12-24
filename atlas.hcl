env "local" {
  src = "file://internal/infrastructure/mariadb/schema/schema.sql"
  url = "maria://user:password@localhost:3306/database"
  dev = "docker://maria/latest/dev"
  format {
      migrate {
        diff = "{{ sql . }}"
      }
    }
}

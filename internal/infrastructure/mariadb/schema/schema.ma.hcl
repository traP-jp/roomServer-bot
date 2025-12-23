table "instances" {
  schema = schema.database
  column "vmid" {
    null     = false
    type     = int
    unsigned = true
  }
  column "user_id" {
    null = false
    type = varchar(100)
  }
  column "template_vmid" {
    null     = false
    type     = int
    unsigned = true
  }
  column "ip_address" {
    null = true
    type = varchar(45)
  }
  column "created_at" {
    null    = true
    type    = timestamp
    default = sql("(current_timestamp())")
  }
  primary_key {
    columns = [column.vmid]
  }
  foreign_key "instances_vm_templates_FK" {
    columns     = [column.template_vmid]
    ref_columns = [table.vm_templates.column.vmid]
    on_update   = CASCADE
    on_delete   = RESTRICT
  }
  index "instances_vm_templates_FK" {
    columns = [column.template_vmid]
  }
}
table "vm_templates" {
  schema = schema.database
  column "vmid" {
    null     = false
    type     = int
    unsigned = true
  }
  column "name" {
    null = false
    type = varchar(100)
  }
  primary_key {
    columns = [column.vmid]
  }
}
schema "database" {
  charset = "utf8mb4"
  collate = "utf8mb4_uca1400_ai_ci"
}

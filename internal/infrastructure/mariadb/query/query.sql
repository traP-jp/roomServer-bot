-- name: GetInstanceByUserID :many
SELECT * FROM instances WHERE user_id = ?;

-- name: CreateInstance :exec
INSERT INTO instances (vmid, user_id, template_vmid, ip_address) VALUES (?, ?, ?, ?);

-- name: DeleteInstanceByVMID :exec
DELETE FROM instances WHERE vmid = ?;

-- name: GetVMTemplateByVMID :one
SELECT * FROM vm_templates WHERE vmid = ?;

-- name: ListVMTemplates :many
SELECT * FROM vm_templates;

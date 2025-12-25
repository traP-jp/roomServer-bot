-- name: GetInstanceByUserID :many
SELECT * FROM instances WHERE user_id = ?;

-- name: GetAllUsers :many
SELECT * FROM users
ORDER BY email;
-- name: SearchEmail :one

SELECT * 
              FROM users 
              WHERE email = $1;
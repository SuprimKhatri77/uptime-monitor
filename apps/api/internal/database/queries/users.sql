-- name: CreateUser :one
INSERT INTO core.users (email, name, avatar_url)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM core.users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM core.users
WHERE email = $1;
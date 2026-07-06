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

-- name: UserExists :one
SELECT EXISTS (SELECT 1 FROM core.users WHERE email = $1);
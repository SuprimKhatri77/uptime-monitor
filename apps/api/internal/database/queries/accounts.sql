-- name: CreateAccount :one
INSERT INTO core.accounts (user_id, provider, provider_account_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAccountByProvider :one
SELECT * FROM core.accounts
WHERE provider = $1 AND provider_account_id = $2;

-- name: GetAccountsByUserID :many
SELECT * FROM core.accounts
WHERE user_id = $1;
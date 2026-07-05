-- name: CreateRefreshToken :one
INSERT INTO core.refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM core.refresh_tokens
WHERE token_hash = $1 AND revoked = false AND expires_at > now();

-- name: RevokeRefreshToken :exec
UPDATE core.refresh_tokens
SET revoked = true
WHERE token_hash = $1;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE core.refresh_tokens
SET revoked = true
WHERE user_id = $1 AND revoked = false;
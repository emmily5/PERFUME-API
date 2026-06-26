-- name: CriarUsuario :one
INSERT INTO users (email, senha_hash)
VALUES ($1, $2)
RETURNING *;

-- name: BuscarUsuarioPorEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: BuscarUsuarioPorID :one
SELECT * FROM users
WHERE id = $1;

-- name: CriarRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expira_em)
VALUES ($1, $2, $3)
RETURNING *;

-- name: BuscarRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token_hash = $1;

-- name: RevogarRefreshToken :exec
UPDATE refresh_tokens
SET revogado = TRUE
WHERE token_hash = $1;

-- name: RevogarTodosTokensDoUsuario :exec
UPDATE refresh_tokens
SET revogado = TRUE
WHERE user_id = $1 AND revogado = FALSE;

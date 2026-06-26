-- Autenticação (Sprint 3): usuários e refresh tokens com rotação.

CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL PRIMARY KEY,
    email      TEXT NOT NULL UNIQUE,
    senha_hash TEXT NOT NULL,
    criado_em  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Refresh tokens: guardamos apenas o HASH (sha-256) do token, nunca o valor em claro.
-- A rotação revoga o token usado e cria um novo a cada refresh.
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expira_em  TIMESTAMPTZ NOT NULL,
    revogado   BOOLEAN NOT NULL DEFAULT FALSE,
    criado_em  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id);

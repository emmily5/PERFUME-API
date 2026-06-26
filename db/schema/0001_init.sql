-- Schema inicial da API de Perfumes (Sprint 2)
-- Relacionamento 1:N -> uma marca possui vários perfumes.

CREATE TABLE IF NOT EXISTS marcas (
    id          BIGSERIAL PRIMARY KEY,
    nome        TEXT NOT NULL UNIQUE,
    pais_origem TEXT NOT NULL DEFAULT '',
    criado_em   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS perfumes (
    id        BIGSERIAL PRIMARY KEY,
    marca_id  BIGINT NOT NULL REFERENCES marcas(id) ON DELETE CASCADE,
    nome      TEXT NOT NULL,
    preco     NUMERIC(10,2) NOT NULL DEFAULT 0,
    tamanho   TEXT NOT NULL DEFAULT '',   -- ex: "50ml", "100ml"
    genero    TEXT NOT NULL DEFAULT '',   -- "masculino", "feminino", "unissex"
    estoque   INT NOT NULL DEFAULT 0,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Índice na chave estrangeira para acelerar a busca de perfumes por marca.
CREATE INDEX IF NOT EXISTS idx_perfumes_marca_id ON perfumes (marca_id);

-- name: CriarPerfume :one
INSERT INTO perfumes (marca_id, nome, preco, tamanho, genero, estoque)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListarPerfumes :many
SELECT * FROM perfumes
ORDER BY id;

-- name: BuscarPerfume :one
SELECT * FROM perfumes
WHERE id = $1;

-- name: ListarPerfumesPorMarca :many
SELECT * FROM perfumes
WHERE marca_id = $1
ORDER BY id;

-- name: AtualizarPerfume :one
UPDATE perfumes
SET marca_id = $2,
    nome = $3,
    preco = $4,
    tamanho = $5,
    genero = $6,
    estoque = $7
WHERE id = $1
RETURNING *;

-- name: DeletarPerfume :exec
DELETE FROM perfumes
WHERE id = $1;

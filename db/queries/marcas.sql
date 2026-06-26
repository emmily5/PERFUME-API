-- name: CriarMarca :one
INSERT INTO marcas (nome, pais_origem)
VALUES ($1, $2)
RETURNING *;

-- name: ListarMarcas :many
SELECT * FROM marcas
ORDER BY nome;

-- name: BuscarMarca :one
SELECT * FROM marcas
WHERE id = $1;

-- name: AtualizarMarca :one
UPDATE marcas
SET nome = $2,
    pais_origem = $3
WHERE id = $1
RETURNING *;

-- name: DeletarMarca :exec
DELETE FROM marcas
WHERE id = $1;

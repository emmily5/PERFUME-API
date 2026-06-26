# 🌸 Perfume API

API REST para uma loja de perfumes, desenvolvida em Go com o router [Chi](https://github.com/go-chi/chi) e persistência em **PostgreSQL** via [sqlc](https://sqlc.dev).

## Tecnologias
- **Go 1.22+**
- **Chi v5** — router HTTP
- **PostgreSQL** — banco de dados relacional
- **pgx/v5** — driver PostgreSQL
- **sqlc** — gera código Go tipado a partir de SQL puro
- **GitHub Actions** — CI/CD automático

---

## Modelagem (relacionamento 1:N)

Uma **marca** possui muitos **perfumes** (1:N). A FK `perfumes.marca_id` referencia
`marcas.id` com `ON DELETE CASCADE`. O endpoint `GET /marcas/{id}` entrega os dados aninhados.

```
marcas (1) ──< perfumes (N)
```

O schema versionado fica em [db/schema/](db/schema/) e as queries SQL em [db/queries/](db/queries/).

---

## Como rodar

```bash
# 1. Dependências
go mod tidy

# 2. Configure a conexão (ajuste usuário/senha/banco)
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/perfumaria?sslmode=disable"

# 3. Aplique o schema e (opcional) o seed de exemplo
make migrate
make seed

# 4. Suba o servidor
make run        # ou: go run ./cmd/api
# Servidor em http://localhost:8080
```

> Para regenerar o código a partir do SQL após alterar `db/queries/` ou `db/schema/`:
> `make sqlc` (precisa do `sqlc` instalado: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`).

---

## Endpoints

### Perfumes (CRUD completo, persistido no banco)

| Método | Rota              | Descrição                     |
|--------|-------------------|-------------------------------|
| GET    | `/perfumes`       | Lista todos os perfumes       |
| GET    | `/perfumes/{id}`  | Busca um perfume pelo ID      |
| POST   | `/perfumes`       | Cria um novo perfume          |
| PUT    | `/perfumes/{id}`  | Atualiza um perfume existente |
| DELETE | `/perfumes/{id}`  | Remove um perfume             |

### Marcas (CRUD + endpoint aninhado 1:N)

| Método | Rota             | Descrição                                       |
|--------|------------------|-------------------------------------------------|
| GET    | `/marcas`        | Lista todas as marcas                           |
| GET    | `/marcas/{id}`   | **Retorna a marca com seus perfumes aninhados** |
| POST   | `/marcas`        | Cria uma nova marca                             |
| PUT    | `/marcas/{id}`   | Atualiza uma marca                              |
| DELETE | `/marcas/{id}`   | Remove a marca (e seus perfumes, via cascade)   |

---

## Exemplos de uso

### Criar uma marca
```bash
curl -X POST http://localhost:8080/marcas \
  -H "Content-Type: application/json" \
  -d '{"nome": "Versace", "pais_origem": "Itália"}'
```

### Criar um perfume vinculado a uma marca
```bash
curl -X POST http://localhost:8080/perfumes \
  -H "Content-Type: application/json" \
  -d '{"marca_id": 1, "nome": "Sauvage", "preco": 650.00, "tamanho": "100ml", "genero": "masculino", "estoque": 10}'
```

### Buscar marca com perfumes aninhados (1:N)
```bash
curl http://localhost:8080/marcas/1
```
```json
{
  "id": 1,
  "nome": "Dior",
  "pais_origem": "França",
  "criado_em": "2026-06-26T08:00:00Z",
  "perfumes": [
    { "id": 1, "marca_id": 1, "nome": "Sauvage", "preco": 650, "tamanho": "100ml", "genero": "masculino", "estoque": 10, "criado_em": "..." }
  ]
}
```

---

## Middlewares

- **Logger** — registra no terminal cada requisição com método, rota, status e tempo de resposta
- **Recovery** — captura `panic` e retorna HTTP 500 sem derrubar o servidor

---

## Estrutura do projeto

```
perfume-api/
├── cmd/api/main.go              # Ponto de entrada (conexão pgx + rotas)
├── db/
│   ├── schema/0001_init.sql     # Schema versionado (marcas + perfumes)
│   ├── queries/                 # Queries SQL puras (entrada do sqlc)
│   │   ├── marcas.sql
│   │   └── perfumes.sql
│   └── seed.sql                 # Dados de exemplo
├── internal/
│   ├── db/                      # Código GERADO pelo sqlc (não editar)
│   ├── handlers/                # Handlers HTTP (perfume.go, marca.go)
│   └── middleware/middleware.go # Logger e Recovery
├── scripts/migrate/main.go      # Aplica db/schema/*.sql e seeds
├── sqlc.yaml                    # Configuração do sqlc
├── Makefile
├── go.mod
└── README.md
```

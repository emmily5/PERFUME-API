# 🌸 Perfume API

API REST para uma loja de perfumes, desenvolvida em Go com o router [Chi](https://github.com/go-chi/chi) e persistência em **PostgreSQL** via [sqlc](https://sqlc.dev).

## Tecnologias
- **Go 1.26+**
- **Chi v5** — router HTTP
- **PostgreSQL** — banco de dados relacional
- **pgx/v5** — driver PostgreSQL
- **sqlc** — gera código Go tipado a partir de SQL puro
- **JWT (golang-jwt/v5)** + **bcrypt** — autenticação e hashing de senha
- **httprate** — rate limiting
- **GraphQL (graphql-go)** — endpoint read-only (bônus)
- **GitHub Actions** — CI com serviço PostgreSQL

---

## Autenticação (Sprint 3)

- `POST /auth/registrar` — cria usuário (senha em bcrypt, mín. 8 caracteres)
- `POST /auth/login` — devolve **access token (JWT, 15 min)** + **refresh token (7 dias)**
- `POST /auth/refresh` — **rotaciona** o refresh token (revoga o antigo, emite um novo);
  reuso de token revogado invalida toda a sessão do usuário
- `POST /auth/logout` — revoga o refresh token

As rotas de **escrita** (`POST`/`PUT`/`DELETE` de perfumes e marcas) exigem
`Authorization: Bearer <access_token>`. As de leitura (`GET`) são públicas.

### Correções de segurança OWASP implementadas
1. **Controle de acesso (A01)** — middleware JWT protege as rotas mutantes
2. **Rate limiting (A07/DoS)** — 100 req/min por IP global; 5 req/min em login/registro
3. **Cabeçalhos de segurança (A05)** — `X-Content-Type-Options`, `X-Frame-Options`, CSP, `Referrer-Policy`
4. **Falhas criptográficas (A02)** — senhas com bcrypt; refresh tokens armazenados só como hash SHA-256
5. **Validação de entrada** — limite de corpo (1 MiB), validação de e-mail/senha, mensagem genérica no login (anti-enumeração)

---

## GraphQL (bônus)

Endpoint read-only em `POST /graphql` (e playground GraphiQL em `GET /graphql`) que entrega o 1:N aninhado:

```graphql
{ marcas { nome pais_origem perfumes { nome preco } } }
{ marca(id: 1) { nome perfumes { nome } } }
```

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

# 2. Configure as variáveis de ambiente (ajuste usuário/senha/banco)
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/perfumaria?sslmode=disable"
export JWT_SECRET="um-segredo-longo-e-aleatorio"

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

🔒 = exige `Authorization: Bearer <access_token>`

### Auth
| Método | Rota               | Descrição                              |
|--------|--------------------|----------------------------------------|
| POST   | `/auth/registrar`  | Cria usuário                           |
| POST   | `/auth/login`      | Login → access + refresh token         |
| POST   | `/auth/refresh`    | Rotaciona o refresh token              |
| POST   | `/auth/logout`     | Revoga o refresh token                 |

### Perfumes (CRUD completo, persistido no banco)

| Método | Rota              | Descrição                        |
|--------|-------------------|----------------------------------|
| GET    | `/perfumes`       | Lista todos os perfumes          |
| GET    | `/perfumes/{id}`  | Busca um perfume pelo ID         |
| POST   | `/perfumes`       | 🔒 Cria um novo perfume          |
| PUT    | `/perfumes/{id}`  | 🔒 Atualiza um perfume existente |
| DELETE | `/perfumes/{id}`  | 🔒 Remove um perfume             |

### Marcas (CRUD + endpoint aninhado 1:N)

| Método | Rota             | Descrição                                       |
|--------|------------------|-------------------------------------------------|
| GET    | `/marcas`        | Lista todas as marcas                           |
| GET    | `/marcas/{id}`   | **Retorna a marca com seus perfumes aninhados** |
| POST   | `/marcas`        | 🔒 Cria uma nova marca                          |
| PUT    | `/marcas/{id}`   | 🔒 Atualiza uma marca                           |
| DELETE | `/marcas/{id}`   | 🔒 Remove a marca (e seus perfumes, via cascade)|

### GraphQL (bônus)
| Método | Rota        | Descrição                                  |
|--------|-------------|--------------------------------------------|
| POST   | `/graphql`  | Consultas (marcas/perfumes aninhados)      |
| GET    | `/graphql`  | Playground GraphiQL                        |

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
├── cmd/api/main.go              # Ponto de entrada (pgx + auth + rotas)
├── db/
│   ├── schema/                  # Schema versionado
│   │   ├── 0001_init.sql        #   marcas + perfumes (1:N)
│   │   └── 0002_auth.sql        #   users + refresh_tokens
│   ├── queries/                 # Queries SQL puras (entrada do sqlc)
│   │   ├── marcas.sql
│   │   ├── perfumes.sql
│   │   └── auth.sql
│   └── seed.sql                 # Dados de exemplo
├── internal/
│   ├── db/                      # Código GERADO pelo sqlc (não editar)
│   ├── auth/                    # JWT, bcrypt, refresh tokens
│   ├── graphqlapi/              # Schema GraphQL (bônus)
│   ├── handlers/                # Handlers HTTP + router
│   │   ├── auth.go              #   registrar/login/refresh/logout
│   │   ├── perfume.go, marca.go
│   │   └── router.go            #   middlewares + rotas
│   └── middleware/              # Logger, Recovery, Auth, Security, RateLimit
├── scripts/migrate/main.go      # Aplica db/schema/*.sql e seeds
├── .github/workflows/ci.yml     # CI com serviço PostgreSQL
├── sqlc.yaml
├── Makefile
└── README.md
```

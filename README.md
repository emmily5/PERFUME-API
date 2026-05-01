# 🌸 Perfume API

API REST para uma loja de perfumes, desenvolvida em Go com o router [Chi](https://github.com/go-chi/chi).

## Tecnologias
- **Go 1.22**
- **Chi v5** — router HTTP
- **Armazenamento em memória** — sem banco de dados
- **GitHub Actions** — CI/CD automático

---

## Como rodar

```bash
# 1. Clone o repositório
git clone https://github.com/seunome/perfume-api.git
cd perfume-api

# 2. Baixe as dependências
go mod tidy

# 3. Inicie o servidor
go run ./cmd/api

# O servidor sobe em http://localhost:8080
```

---

## Endpoints

| Método | Rota              | Descrição                     |
|--------|-------------------|-------------------------------|
| GET    | `/`               | Boas-vindas da API            |
| GET    | `/perfumes`       | Lista todos os perfumes       |
| GET    | `/perfumes/{id}`  | Busca um perfume pelo ID      |
| POST   | `/perfumes`       | Cria um novo perfume          |
| PUT    | `/perfumes/{id}`  | Atualiza um perfume existente |
| DELETE | `/perfumes/{id}`  | Remove um perfume             |

---

## Exemplos de uso

### Listar perfumes
```bash
curl http://localhost:8080/perfumes
```

### Buscar por ID
```bash
curl http://localhost:8080/perfumes/1
```

### Criar perfume
```bash
curl -X POST http://localhost:8080/perfumes \
  -H "Content-Type: application/json" \
  -d '{
    "nome": "Bleu de Chanel",
    "marca": "Chanel",
    "preco": 720.00,
    "tamanho": "100ml",
    "genero": "masculino",
    "estoque": 5
  }'
```

### Atualizar perfume
```bash
curl -X PUT http://localhost:8080/perfumes/1 \
  -H "Content-Type: application/json" \
  -d '{
    "nome": "Sauvage Elixir",
    "marca": "Dior",
    "preco": 850.00,
    "tamanho": "60ml",
    "genero": "masculino",
    "estoque": 3
  }'
```

### Deletar perfume
```bash
curl -X DELETE http://localhost:8080/perfumes/1
```

---

## Middlewares

- **Logger** — registra no terminal cada requisição com método, rota, status e tempo de resposta
- **Recovery** — captura `panic` e retorna HTTP 500 sem derrubar o servidor

---

## Testes

```bash
go test ./... -v
```

---

## Estrutura do projeto

```
perfume-api/
├── cmd/
│   └── api/
│       └── main.go          # Ponto de entrada
├── internal/
│   ├── handlers/
│   │   ├── perfume.go       # Handlers HTTP
│   │   └── perfume_test.go  # Testes dos handlers
│   ├── middleware/
│   │   └── middleware.go    # Logger e Recovery
│   ├── models/
│   │   └── perfume.go       # Struct Perfume
│   └── store/
│       ├── store.go         # Armazenamento em memória
│       └── store_test.go    # Testes do store
├── .github/
│   └── workflows/
│       └── ci.yml           # GitHub Actions
├── go.mod
└── README.md
```

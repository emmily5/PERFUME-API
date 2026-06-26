.PHONY: sqlc migrate seed run tidy

# Gera o código Go tipado a partir das queries SQL.
sqlc:
	sqlc generate

# Aplica o schema versionado (db/schema/*.sql) no banco de DATABASE_URL.
migrate:
	go run ./scripts/migrate

# Carrega dados de exemplo (db/seed.sql).
seed:
	go run ./scripts/migrate -seed db/seed.sql

# Sobe a API (precisa de DATABASE_URL no ambiente).
run:
	go run ./cmd/api

tidy:
	go mod tidy

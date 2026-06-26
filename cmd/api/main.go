package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seunome/perfume-api/internal/db"
	"github.com/seunome/perfume-api/internal/handlers"
)

func main() {
	// String de conexão lida do ambiente.
	// Ex: postgres://usuario:senha@localhost:5432/perfumaria?sslmode=disable
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("variável de ambiente DATABASE_URL não definida")
	}

	// Conecta ao PostgreSQL com um pool de conexões.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("erro ao criar pool de conexões: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("não foi possível conectar ao banco: %v", err)
	}
	log.Println("✅ Conectado ao PostgreSQL")

	// Queries tipadas geradas pelo sqlc + router com middlewares e rotas.
	queries := db.New(pool)
	r := handlers.NewRouter(queries)

	log.Println("🌸 Servidor rodando em http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

// Migrator simples: aplica, em ordem, todos os arquivos .sql de db/schema/.
// Uso: DATABASE_URL=postgres://... go run ./scripts/migrate
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5"
)

func main() {
	// -seed aplica um único arquivo .sql (ex: db/seed.sql) em vez do schema.
	seed := flag.String("seed", "", "caminho de um arquivo .sql avulso para aplicar")
	flag.Parse()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL não definida")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("erro ao conectar: %v", err)
	}
	defer conn.Close(ctx)

	var arquivos []string
	if *seed != "" {
		arquivos = []string{*seed}
	} else {
		arquivos, err = filepath.Glob("db/schema/*.sql")
		if err != nil {
			log.Fatalf("erro ao listar migrations: %v", err)
		}
		sort.Strings(arquivos)
	}

	for _, arq := range arquivos {
		sqlBytes, err := os.ReadFile(arq)
		if err != nil {
			log.Fatalf("erro ao ler %s: %v", arq, err)
		}
		if _, err := conn.Exec(ctx, string(sqlBytes)); err != nil {
			log.Fatalf("erro ao aplicar %s: %v", arq, err)
		}
		log.Printf("aplicado: %s", arq)
	}
	log.Println("✅ schema aplicado com sucesso")
}

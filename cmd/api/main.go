package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/seunome/perfume-api/internal/handlers"
	"github.com/seunome/perfume-api/internal/middleware"
	"github.com/seunome/perfume-api/internal/store"
)

func main() {
	// Inicializa o store em memória com dados de exemplo
	s := store.New()

	// Inicializa os handlers
	h := handlers.NewPerfumeHandler(s)

	// Cria o router Chi
	r := chi.NewRouter()

	// Middlewares globais (ordem importa: recovery primeiro para capturar panics)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)

	// Rotas
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"mensagem":"Bem-vindo à API da Loja de Perfumes 🌸","versao":"1.0"}`))
	})

	r.Route("/perfumes", func(r chi.Router) {
		r.Get("/", h.ListarPerfumes)       // GET  /perfumes
		r.Post("/", h.CriarPerfume)        // POST /perfumes
		r.Get("/{id}", h.BuscarPerfume)    // GET  /perfumes/{id}
		r.Put("/{id}", h.AtualizarPerfume) // PUT  /perfumes/{id}
		r.Delete("/{id}", h.DeletarPerfume) // DELETE /perfumes/{id}
	})

	log.Println("🌸 Servidor rodando em http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

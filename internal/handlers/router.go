package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/seunome/perfume-api/internal/db"
	"github.com/seunome/perfume-api/internal/middleware"
)

// NewRouter monta o router Chi com middlewares e todas as rotas da API.
// É usado tanto pelo main quanto pelos testes de integração.
func NewRouter(q *db.Queries) http.Handler {
	h := New(q)

	r := chi.NewRouter()
	// Recovery primeiro para capturar panics; depois o Logger.
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"mensagem":"Bem-vindo à API da Loja de Perfumes 🌸","versao":"2.0"}`))
	})

	// CRUD de perfumes (entidade persistida no banco).
	r.Route("/perfumes", func(r chi.Router) {
		r.Get("/", h.ListarPerfumes)
		r.Post("/", h.CriarPerfume)
		r.Get("/{id}", h.BuscarPerfume)
		r.Put("/{id}", h.AtualizarPerfume)
		r.Delete("/{id}", h.DeletarPerfume)
	})

	// CRUD de marcas + endpoint aninhado (1:N) em GET /marcas/{id}.
	r.Route("/marcas", func(r chi.Router) {
		r.Get("/", h.ListarMarcas)
		r.Post("/", h.CriarMarca)
		r.Get("/{id}", h.BuscarMarca) // retorna a marca com seus perfumes aninhados
		r.Put("/{id}", h.AtualizarMarca)
		r.Delete("/{id}", h.DeletarMarca)
	})

	return r
}

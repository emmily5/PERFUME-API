package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/seunome/perfume-api/internal/auth"
	"github.com/seunome/perfume-api/internal/db"
	"github.com/seunome/perfume-api/internal/graphqlapi"
	"github.com/seunome/perfume-api/internal/middleware"
)

// NewRouter monta o router Chi com middlewares, segurança e todas as rotas.
// É usado tanto pelo main quanto pelos testes de integração.
func NewRouter(q *db.Queries, authSvc *auth.Service) (http.Handler, error) {
	h := New(q, authSvc)

	gql, err := graphqlapi.NewHandler(q)
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()

	// --- Middlewares globais ---
	r.Use(middleware.Recovery)                          // captura panics
	r.Use(middleware.Logger)                            // log de requisições
	r.Use(middleware.SecurityHeaders)                   // OWASP A05: cabeçalhos de segurança
	r.Use(middleware.LimitarTamanhoBody(1 << 20))       // limita corpo a 1 MiB
	r.Use(httprate.LimitByIP(100, time.Minute))         // OWASP: rate limit global por IP

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"mensagem":"Bem-vindo à API da Loja de Perfumes 🌸","versao":"3.0"}`))
	})

	// --- Autenticação (público) ---
	r.Route("/auth", func(r chi.Router) {
		// Rate limit mais agressivo no login para mitigar força bruta (OWASP A07).
		r.With(httprate.LimitByIP(5, time.Minute)).Post("/login", h.Login)
		r.With(httprate.LimitByIP(5, time.Minute)).Post("/registrar", h.Registrar)
		r.Post("/refresh", h.Refresh)
		r.Post("/logout", h.Logout)
	})

	// --- GraphQL (bônus, read-only, público) ---
	r.Handle("/graphql", gql)

	// --- Perfumes ---
	r.Route("/perfumes", func(r chi.Router) {
		// Leitura pública.
		r.Get("/", h.ListarPerfumes)
		r.Get("/{id}", h.BuscarPerfume)

		// Escrita protegida por JWT (OWASP A01: controle de acesso).
		r.Group(func(r chi.Router) {
			r.Use(middleware.Autenticacao(authSvc))
			r.Post("/", h.CriarPerfume)
			r.Put("/{id}", h.AtualizarPerfume)
			r.Delete("/{id}", h.DeletarPerfume)
		})
	})

	// --- Marcas (com endpoint aninhado 1:N em GET /marcas/{id}) ---
	r.Route("/marcas", func(r chi.Router) {
		r.Get("/", h.ListarMarcas)
		r.Get("/{id}", h.BuscarMarca)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Autenticacao(authSvc))
			r.Post("/", h.CriarMarca)
			r.Put("/{id}", h.AtualizarMarca)
			r.Delete("/{id}", h.DeletarMarca)
		})
	})

	return r, nil
}

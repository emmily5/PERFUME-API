package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/seunome/perfume-api/internal/db"
)

// Handler agrupa as dependências dos handlers HTTP.
// Recebe o *db.Queries gerado pelo sqlc, que executa as queries no PostgreSQL.
type Handler struct {
	q *db.Queries
}

// New cria um Handler com o acesso ao banco injetado.
func New(q *db.Queries) *Handler {
	return &Handler{q: q}
}

// --- helpers ---

func respJSON(w http.ResponseWriter, status int, dado any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dado)
}

func respErro(w http.ResponseWriter, status int, msg string) {
	respJSON(w, status, map[string]string{"erro": msg})
}

// idFromURL extrai o parâmetro {id} da URL como int64 (BIGSERIAL no banco).
func idFromURL(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}

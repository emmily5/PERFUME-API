package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/seunome/perfume-api/internal/db"
)

// ListarPerfumes GET /perfumes — lista todos os perfumes.
func (h *Handler) ListarPerfumes(w http.ResponseWriter, r *http.Request) {
	perfumes, err := h.q.ListarPerfumes(r.Context())
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao listar perfumes")
		return
	}
	if perfumes == nil {
		perfumes = []db.Perfume{}
	}
	respJSON(w, http.StatusOK, perfumes)
}

// BuscarPerfume GET /perfumes/{id} — retorna um perfume pelo ID.
func (h *Handler) BuscarPerfume(w http.ResponseWriter, r *http.Request) {
	id, err := idFromURL(r)
	if err != nil {
		respErro(w, http.StatusBadRequest, "id inválido")
		return
	}

	perfume, err := h.q.BuscarPerfume(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		respErro(w, http.StatusNotFound, "perfume não encontrado")
		return
	}
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao buscar perfume")
		return
	}
	respJSON(w, http.StatusOK, perfume)
}

// CriarPerfume POST /perfumes — cria um perfume vinculado a uma marca existente.
func (h *Handler) CriarPerfume(w http.ResponseWriter, r *http.Request) {
	var in db.CriarPerfumeParams
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respErro(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if in.Nome == "" {
		respErro(w, http.StatusBadRequest, "nome é obrigatório")
		return
	}
	if in.MarcaID == 0 {
		respErro(w, http.StatusBadRequest, "marca_id é obrigatório")
		return
	}

	perfume, err := h.q.CriarPerfume(r.Context(), in)
	if violaChaveEstrangeira(err) {
		respErro(w, http.StatusBadRequest, "marca_id não corresponde a uma marca existente")
		return
	}
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao criar perfume")
		return
	}
	respJSON(w, http.StatusCreated, perfume)
}

// AtualizarPerfume PUT /perfumes/{id} — substitui os dados de um perfume.
func (h *Handler) AtualizarPerfume(w http.ResponseWriter, r *http.Request) {
	id, err := idFromURL(r)
	if err != nil {
		respErro(w, http.StatusBadRequest, "id inválido")
		return
	}

	var in db.AtualizarPerfumeParams
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respErro(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	in.ID = id

	if in.Nome == "" {
		respErro(w, http.StatusBadRequest, "nome é obrigatório")
		return
	}
	if in.MarcaID == 0 {
		respErro(w, http.StatusBadRequest, "marca_id é obrigatório")
		return
	}

	perfume, err := h.q.AtualizarPerfume(r.Context(), in)
	if errors.Is(err, pgx.ErrNoRows) {
		respErro(w, http.StatusNotFound, "perfume não encontrado")
		return
	}
	if violaChaveEstrangeira(err) {
		respErro(w, http.StatusBadRequest, "marca_id não corresponde a uma marca existente")
		return
	}
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao atualizar perfume")
		return
	}
	respJSON(w, http.StatusOK, perfume)
}

// DeletarPerfume DELETE /perfumes/{id} — remove um perfume.
func (h *Handler) DeletarPerfume(w http.ResponseWriter, r *http.Request) {
	id, err := idFromURL(r)
	if err != nil {
		respErro(w, http.StatusBadRequest, "id inválido")
		return
	}

	// Confere se existe para responder 404 quando for o caso (DELETE é idempotente no SQL).
	if _, err := h.q.BuscarPerfume(r.Context(), id); errors.Is(err, pgx.ErrNoRows) {
		respErro(w, http.StatusNotFound, "perfume não encontrado")
		return
	}

	if err := h.q.DeletarPerfume(r.Context(), id); err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao deletar perfume")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// violaChaveEstrangeira detecta o erro 23503 do PostgreSQL (foreign_key_violation).
func violaChaveEstrangeira(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

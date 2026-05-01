package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/seunome/perfume-api/internal/models"
	"github.com/seunome/perfume-api/internal/store"
)

// PerfumeHandler agrupa todos os handlers relacionados a perfumes
type PerfumeHandler struct {
	store *store.PerfumeStore
}

// NewPerfumeHandler cria um PerfumeHandler com o store injetado
func NewPerfumeHandler(s *store.PerfumeStore) *PerfumeHandler {
	return &PerfumeHandler{store: s}
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

func idFromURL(r *http.Request) (int, error) {
	return strconv.Atoi(chi.URLParam(r, "id"))
}

// --- handlers ---

// ListarPerfumes GET /perfumes
// Retorna todos os perfumes cadastrados
func (h *PerfumeHandler) ListarPerfumes(w http.ResponseWriter, r *http.Request) {
	perfumes := h.store.Listar()
	respJSON(w, http.StatusOK, perfumes)
}

// BuscarPerfume GET /perfumes/{id}
// Retorna um perfume específico pelo ID
func (h *PerfumeHandler) BuscarPerfume(w http.ResponseWriter, r *http.Request) {
	id, err := idFromURL(r)
	if err != nil {
		respErro(w, http.StatusBadRequest, "id inválido")
		return
	}

	perfume, err := h.store.BuscarPorID(id)
	if errors.Is(err, store.ErrNaoEncontrado) {
		respErro(w, http.StatusNotFound, "perfume não encontrado")
		return
	}

	respJSON(w, http.StatusOK, perfume)
}

// CriarPerfume POST /perfumes
// Cria um novo perfume a partir do corpo JSON da requisição
func (h *PerfumeHandler) CriarPerfume(w http.ResponseWriter, r *http.Request) {
	var p models.Perfume
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respErro(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if p.Nome == "" || p.Marca == "" {
		respErro(w, http.StatusBadRequest, "nome e marca são obrigatórios")
		return
	}

	criado := h.store.Criar(p)
	respJSON(w, http.StatusCreated, criado)
}

// AtualizarPerfume PUT /perfumes/{id}
// Substitui completamente os dados de um perfume
func (h *PerfumeHandler) AtualizarPerfume(w http.ResponseWriter, r *http.Request) {
	id, err := idFromURL(r)
	if err != nil {
		respErro(w, http.StatusBadRequest, "id inválido")
		return
	}

	var p models.Perfume
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respErro(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	atualizado, err := h.store.Atualizar(id, p)
	if errors.Is(err, store.ErrNaoEncontrado) {
		respErro(w, http.StatusNotFound, "perfume não encontrado")
		return
	}

	respJSON(w, http.StatusOK, atualizado)
}

// DeletarPerfume DELETE /perfumes/{id}
// Remove um perfume pelo ID
func (h *PerfumeHandler) DeletarPerfume(w http.ResponseWriter, r *http.Request) {
	id, err := idFromURL(r)
	if err != nil {
		respErro(w, http.StatusBadRequest, "id inválido")
		return
	}

	if err := h.store.Deletar(id); errors.Is(err, store.ErrNaoEncontrado) {
		respErro(w, http.StatusNotFound, "perfume não encontrado")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

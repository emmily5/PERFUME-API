package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/seunome/perfume-api/internal/db"
)

// marcaComPerfumes é a resposta aninhada do relacionamento 1:N.
// Embute a Marca e adiciona a lista de perfumes que pertencem a ela.
type marcaComPerfumes struct {
	db.Marca
	Perfumes []db.Perfume `json:"perfumes"`
}

// ListarMarcas GET /marcas — lista todas as marcas.
func (h *Handler) ListarMarcas(w http.ResponseWriter, r *http.Request) {
	marcas, err := h.q.ListarMarcas(r.Context())
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao listar marcas")
		return
	}
	if marcas == nil {
		marcas = []db.Marca{}
	}
	respJSON(w, http.StatusOK, marcas)
}

// BuscarMarca GET /marcas/{id} — retorna a marca com seus perfumes aninhados (1:N).
func (h *Handler) BuscarMarca(w http.ResponseWriter, r *http.Request) {
	id, err := idFromURL(r)
	if err != nil {
		respErro(w, http.StatusBadRequest, "id inválido")
		return
	}

	marca, err := h.q.BuscarMarca(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		respErro(w, http.StatusNotFound, "marca não encontrada")
		return
	}
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao buscar marca")
		return
	}

	perfumes, err := h.q.ListarPerfumesPorMarca(r.Context(), id)
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao buscar perfumes da marca")
		return
	}
	if perfumes == nil {
		perfumes = []db.Perfume{}
	}

	respJSON(w, http.StatusOK, marcaComPerfumes{Marca: marca, Perfumes: perfumes})
}

// CriarMarca POST /marcas — cria uma nova marca.
func (h *Handler) CriarMarca(w http.ResponseWriter, r *http.Request) {
	var in db.CriarMarcaParams
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respErro(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if in.Nome == "" {
		respErro(w, http.StatusBadRequest, "nome é obrigatório")
		return
	}

	marca, err := h.q.CriarMarca(r.Context(), in)
	if violaUnico(err) {
		respErro(w, http.StatusConflict, "já existe uma marca com esse nome")
		return
	}
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao criar marca")
		return
	}
	respJSON(w, http.StatusCreated, marca)
}

// AtualizarMarca PUT /marcas/{id} — atualiza os dados de uma marca.
func (h *Handler) AtualizarMarca(w http.ResponseWriter, r *http.Request) {
	id, err := idFromURL(r)
	if err != nil {
		respErro(w, http.StatusBadRequest, "id inválido")
		return
	}

	var in db.AtualizarMarcaParams
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respErro(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	in.ID = id
	if in.Nome == "" {
		respErro(w, http.StatusBadRequest, "nome é obrigatório")
		return
	}

	marca, err := h.q.AtualizarMarca(r.Context(), in)
	if errors.Is(err, pgx.ErrNoRows) {
		respErro(w, http.StatusNotFound, "marca não encontrada")
		return
	}
	if violaUnico(err) {
		respErro(w, http.StatusConflict, "já existe uma marca com esse nome")
		return
	}
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao atualizar marca")
		return
	}
	respJSON(w, http.StatusOK, marca)
}

// DeletarMarca DELETE /marcas/{id} — remove uma marca (e seus perfumes, via ON DELETE CASCADE).
func (h *Handler) DeletarMarca(w http.ResponseWriter, r *http.Request) {
	id, err := idFromURL(r)
	if err != nil {
		respErro(w, http.StatusBadRequest, "id inválido")
		return
	}

	if _, err := h.q.BuscarMarca(r.Context(), id); errors.Is(err, pgx.ErrNoRows) {
		respErro(w, http.StatusNotFound, "marca não encontrada")
		return
	}

	if err := h.q.DeletarMarca(r.Context(), id); err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao deletar marca")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// violaUnico detecta o erro 23505 do PostgreSQL (unique_violation).
func violaUnico(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

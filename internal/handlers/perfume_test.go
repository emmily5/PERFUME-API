package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/seunome/perfume-api/internal/handlers"
	"github.com/seunome/perfume-api/internal/models"
	"github.com/seunome/perfume-api/internal/store"
)

// novoRouter cria um router de teste já com as rotas configuradas
func novoRouter() http.Handler {
	s := store.New()
	h := handlers.NewPerfumeHandler(s)

	r := chi.NewRouter()
	r.Get("/perfumes", h.ListarPerfumes)
	r.Post("/perfumes", h.CriarPerfume)
	r.Get("/perfumes/{id}", h.BuscarPerfume)
	r.Put("/perfumes/{id}", h.AtualizarPerfume)
	r.Delete("/perfumes/{id}", h.DeletarPerfume)
	return r
}

func TestListarPerfumesHTTP(t *testing.T) {
	r := novoRouter()

	req := httptest.NewRequest(http.MethodGet, "/perfumes", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("esperava status 200, recebeu %d", rr.Code)
	}

	var perfumes []models.Perfume
	if err := json.NewDecoder(rr.Body).Decode(&perfumes); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}
	if len(perfumes) == 0 {
		t.Error("esperava pelo menos um perfume na lista")
	}
}

func TestCriarPerfumeHTTP(t *testing.T) {
	r := novoRouter()

	body := `{"nome":"Gabrielle","marca":"Chanel","preco":590.00,"tamanho":"50ml","genero":"feminino","estoque":7}`
	req := httptest.NewRequest(http.MethodPost, "/perfumes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("esperava status 201, recebeu %d", rr.Code)
	}

	var criado models.Perfume
	json.NewDecoder(rr.Body).Decode(&criado)
	if criado.ID == 0 {
		t.Error("perfume criado deveria ter um ID")
	}
	if criado.Nome != "Gabrielle" {
		t.Errorf("nome esperado Gabrielle, recebeu %q", criado.Nome)
	}
}

func TestBuscarPerfumeHTTP(t *testing.T) {
	r := novoRouter()

	req := httptest.NewRequest(http.MethodGet, "/perfumes/1", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("esperava status 200, recebeu %d", rr.Code)
	}
}

func TestBuscarPerfumeNaoEncontrado(t *testing.T) {
	r := novoRouter()

	req := httptest.NewRequest(http.MethodGet, "/perfumes/9999", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("esperava status 404, recebeu %d", rr.Code)
	}
}

func TestDeletarPerfumeHTTP(t *testing.T) {
	r := novoRouter()

	req := httptest.NewRequest(http.MethodDelete, "/perfumes/1", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("esperava status 204, recebeu %d", rr.Code)
	}
}

func TestCriarPerfumeSemNome(t *testing.T) {
	r := novoRouter()

	body := `{"marca":"Dior","preco":300.00}`
	req := httptest.NewRequest(http.MethodPost, "/perfumes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("esperava status 400 para nome vazio, recebeu %d", rr.Code)
	}
}

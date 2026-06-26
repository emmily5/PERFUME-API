package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seunome/perfume-api/internal/db"
	"github.com/seunome/perfume-api/internal/handlers"
)

// testPool é o pool compartilhado pelos testes de integração.
// Fica nil quando não há banco configurado, fazendo os testes darem skip.
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn != "" {
		pool, err := pgxpool.New(context.Background(), dsn)
		if err == nil && pool.Ping(context.Background()) == nil {
			testPool = pool
			defer pool.Close()
		}
	}
	os.Exit(m.Run())
}

// setup abre uma transação isolada e devolve um router que opera dentro dela.
// O cleanup faz rollback, então nenhum teste persiste dados no banco.
func setup(t *testing.T) (http.Handler, func()) {
	t.Helper()
	if testPool == nil {
		t.Skip("defina DATABASE_URL (ou TEST_DATABASE_URL) para rodar os testes de integração")
	}
	tx, err := testPool.Begin(context.Background())
	if err != nil {
		t.Fatalf("erro ao iniciar transação: %v", err)
	}
	router := handlers.NewRouter(db.New(tx))
	cleanup := func() { _ = tx.Rollback(context.Background()) }
	return router, cleanup
}

// do executa uma requisição contra o router e devolve o response recorder.
func do(t *testing.T, router http.Handler, metodo, rota string, corpo any) *httptest.ResponseRecorder {
	t.Helper()
	var body *bytes.Reader
	if corpo != nil {
		b, _ := json.Marshal(corpo)
		body = bytes.NewReader(b)
	} else {
		body = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(metodo, rota, body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("erro ao decodificar resposta %q: %v", rec.Body.String(), err)
	}
	return out
}

// criarMarca é um helper que cria uma marca via API e devolve seu ID.
func criarMarca(t *testing.T, router http.Handler, nome string) int64 {
	t.Helper()
	rec := do(t, router, http.MethodPost, "/marcas", map[string]any{"nome": nome})
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar marca: esperava 201, veio %d (%s)", rec.Code, rec.Body.String())
	}
	return decode[db.Marca](t, rec).ID
}

func TestCRUDPerfume(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	marcaID := criarMarca(t, router, "Marca Teste CRUD")

	// CREATE
	rec := do(t, router, http.MethodPost, "/perfumes", map[string]any{
		"marca_id": marcaID, "nome": "Teste", "preco": 199.90,
		"tamanho": "50ml", "genero": "unissex", "estoque": 3,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("CREATE: esperava 201, veio %d (%s)", rec.Code, rec.Body.String())
	}
	criado := decode[db.Perfume](t, rec)
	if criado.ID == 0 || criado.Preco != 199.90 {
		t.Fatalf("CREATE: perfume inesperado: %+v", criado)
	}

	// READ
	rec = do(t, router, http.MethodGet, "/perfumes/"+itoa(criado.ID), nil)
	if rec.Code != http.StatusOK || decode[db.Perfume](t, rec).Nome != "Teste" {
		t.Fatalf("READ: resposta inesperada %d (%s)", rec.Code, rec.Body.String())
	}

	// UPDATE
	rec = do(t, router, http.MethodPut, "/perfumes/"+itoa(criado.ID), map[string]any{
		"marca_id": marcaID, "nome": "Teste Atualizado", "preco": 250.00,
		"tamanho": "100ml", "genero": "unissex", "estoque": 1,
	})
	if rec.Code != http.StatusOK || decode[db.Perfume](t, rec).Nome != "Teste Atualizado" {
		t.Fatalf("UPDATE: resposta inesperada %d (%s)", rec.Code, rec.Body.String())
	}

	// DELETE
	rec = do(t, router, http.MethodDelete, "/perfumes/"+itoa(criado.ID), nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE: esperava 204, veio %d", rec.Code)
	}
	rec = do(t, router, http.MethodGet, "/perfumes/"+itoa(criado.ID), nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("READ após DELETE: esperava 404, veio %d", rec.Code)
	}
}

func TestMarcaComPerfumesAninhados(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	marcaID := criarMarca(t, router, "Marca Teste Aninhado")
	for _, nome := range []string{"P1", "P2"} {
		rec := do(t, router, http.MethodPost, "/perfumes", map[string]any{
			"marca_id": marcaID, "nome": nome, "preco": 100,
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("criar %s: %d (%s)", nome, rec.Code, rec.Body.String())
		}
	}

	rec := do(t, router, http.MethodGet, "/marcas/"+itoa(marcaID), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET marca: esperava 200, veio %d", rec.Code)
	}
	type marcaComPerfumes struct {
		ID       int64        `json:"id"`
		Perfumes []db.Perfume `json:"perfumes"`
	}
	out := decode[marcaComPerfumes](t, rec)
	if out.ID != marcaID {
		t.Fatalf("id da marca inesperado: %d", out.ID)
	}
	if len(out.Perfumes) != 2 {
		t.Fatalf("esperava 2 perfumes aninhados, veio %d", len(out.Perfumes))
	}
}

func TestPerfumeMarcaInexistente(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	rec := do(t, router, http.MethodPost, "/perfumes", map[string]any{
		"marca_id": 999999999, "nome": "Fantasma", "preco": 1,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400 para marca_id inexistente, veio %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestPerfumeNaoEncontrado(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	rec := do(t, router, http.MethodGet, "/perfumes/999999999", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404, veio %d", rec.Code)
	}
}

func TestMarcaNomeDuplicado(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	criarMarca(t, router, "Marca Duplicada Teste")
	rec := do(t, router, http.MethodPost, "/marcas", map[string]any{"nome": "Marca Duplicada Teste"})
	if rec.Code != http.StatusConflict {
		t.Fatalf("esperava 409 para nome duplicado, veio %d (%s)", rec.Code, rec.Body.String())
	}
}

// itoa converte int64 para string para montar as URLs nos testes.
func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

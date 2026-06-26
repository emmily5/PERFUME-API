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
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seunome/perfume-api/internal/auth"
	"github.com/seunome/perfume-api/internal/db"
	"github.com/seunome/perfume-api/internal/handlers"
)

// testPool é o pool compartilhado pelos testes de integração.
// Fica nil quando não há banco configurado, fazendo os testes darem skip.
var testPool *pgxpool.Pool

const testSecret = "segredo-de-teste-0123456789-abcdef"

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
	authSvc := auth.NewService(testSecret, 15*time.Minute, 7*24*time.Hour)
	router, err := handlers.NewRouter(db.New(tx), authSvc)
	if err != nil {
		t.Fatalf("erro ao montar router: %v", err)
	}
	cleanup := func() { _ = tx.Rollback(context.Background()) }
	return router, cleanup
}

// do executa uma requisição. Se token != "", envia o header Authorization.
func do(t *testing.T, router http.Handler, metodo, rota string, corpo any, token string) *httptest.ResponseRecorder {
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
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
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

type tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// autenticar registra um usuário e faz login, devolvendo os tokens emitidos.
func autenticar(t *testing.T, router http.Handler, email string) tokens {
	t.Helper()
	rec := do(t, router, http.MethodPost, "/auth/registrar", map[string]any{
		"email": email, "senha": "senhaSegura123",
	}, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("registrar: esperava 201, veio %d (%s)", rec.Code, rec.Body.String())
	}
	rec = do(t, router, http.MethodPost, "/auth/login", map[string]any{
		"email": email, "senha": "senhaSegura123",
	}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("login: esperava 200, veio %d (%s)", rec.Code, rec.Body.String())
	}
	return decode[tokens](t, rec)
}

func criarMarca(t *testing.T, router http.Handler, nome, token string) int64 {
	t.Helper()
	rec := do(t, router, http.MethodPost, "/marcas", map[string]any{"nome": nome}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar marca: esperava 201, veio %d (%s)", rec.Code, rec.Body.String())
	}
	return decode[db.Marca](t, rec).ID
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

// ---------------------- Autenticação / Segurança ----------------------

func TestRegistrarELogin(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	tk := autenticar(t, router, "alice@teste.com")
	if tk.AccessToken == "" || tk.RefreshToken == "" {
		t.Fatalf("esperava access e refresh token preenchidos: %+v", tk)
	}
}

func TestLoginSenhaErrada(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	autenticar(t, router, "bob@teste.com")
	rec := do(t, router, http.MethodPost, "/auth/login", map[string]any{
		"email": "bob@teste.com", "senha": "errada-errada",
	}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401 para senha errada, veio %d", rec.Code)
	}
}

func TestLoginEmailInexistente(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	rec := do(t, router, http.MethodPost, "/auth/login", map[string]any{
		"email": "ninguem@teste.com", "senha": "qualquer123",
	}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401 para e-mail inexistente, veio %d", rec.Code)
	}
}

func TestRegistrarEmailDuplicado(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	body := map[string]any{"email": "dup@teste.com", "senha": "senhaSegura123"}
	if rec := do(t, router, http.MethodPost, "/auth/registrar", body, ""); rec.Code != http.StatusCreated {
		t.Fatalf("primeiro registro: esperava 201, veio %d", rec.Code)
	}
	if rec := do(t, router, http.MethodPost, "/auth/registrar", body, ""); rec.Code != http.StatusConflict {
		t.Fatalf("segundo registro: esperava 409, veio %d", rec.Code)
	}
}

func TestRegistrarSenhaCurta(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	rec := do(t, router, http.MethodPost, "/auth/registrar", map[string]any{
		"email": "curta@teste.com", "senha": "123",
	}, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400 para senha curta, veio %d", rec.Code)
	}
}

func TestRotaProtegidaSemToken(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	rec := do(t, router, http.MethodPost, "/perfumes", map[string]any{
		"marca_id": 1, "nome": "Sem Auth", "preco": 100,
	}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401 sem token, veio %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestRotaProtegidaComTokenInvalido(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	rec := do(t, router, http.MethodPost, "/marcas", map[string]any{"nome": "X"}, "token.invalido.aqui")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401 com token inválido, veio %d", rec.Code)
	}
}

func TestRefreshRotaciona(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	tk := autenticar(t, router, "rot@teste.com")
	rec := do(t, router, http.MethodPost, "/auth/refresh", map[string]any{
		"refresh_token": tk.RefreshToken,
	}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh: esperava 200, veio %d (%s)", rec.Code, rec.Body.String())
	}
	novo := decode[tokens](t, rec)
	if novo.RefreshToken == "" || novo.RefreshToken == tk.RefreshToken {
		t.Fatalf("rotação deveria gerar um refresh token diferente")
	}
}

func TestRefreshReusoRejeitado(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	tk := autenticar(t, router, "reuso@teste.com")
	// Primeiro uso: rotaciona com sucesso.
	if rec := do(t, router, http.MethodPost, "/auth/refresh", map[string]any{
		"refresh_token": tk.RefreshToken,
	}, ""); rec.Code != http.StatusOK {
		t.Fatalf("primeiro refresh: esperava 200, veio %d", rec.Code)
	}
	// Reuso do token já rotacionado: deve ser rejeitado.
	rec := do(t, router, http.MethodPost, "/auth/refresh", map[string]any{
		"refresh_token": tk.RefreshToken,
	}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("reuso de refresh: esperava 401, veio %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestSecurityHeaders(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	rec := do(t, router, http.MethodGet, "/", nil, "")
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("esperava X-Content-Type-Options=nosniff, veio %q", got)
	}
	if got := rec.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("esperava X-Frame-Options=DENY, veio %q", got)
	}
}

// ---------------------- CRUD + relacionamento 1:N ----------------------

func TestCRUDPerfume(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	tk := autenticar(t, router, "crud@teste.com")
	marcaID := criarMarca(t, router, "Marca Teste CRUD", tk.AccessToken)

	// CREATE (protegido)
	rec := do(t, router, http.MethodPost, "/perfumes", map[string]any{
		"marca_id": marcaID, "nome": "Teste", "preco": 199.90,
		"tamanho": "50ml", "genero": "unissex", "estoque": 3,
	}, tk.AccessToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("CREATE: esperava 201, veio %d (%s)", rec.Code, rec.Body.String())
	}
	criado := decode[db.Perfume](t, rec)
	if criado.ID == 0 || criado.Preco != 199.90 {
		t.Fatalf("CREATE: perfume inesperado: %+v", criado)
	}

	// READ (público)
	rec = do(t, router, http.MethodGet, "/perfumes/"+itoa(criado.ID), nil, "")
	if rec.Code != http.StatusOK || decode[db.Perfume](t, rec).Nome != "Teste" {
		t.Fatalf("READ: resposta inesperada %d (%s)", rec.Code, rec.Body.String())
	}

	// UPDATE (protegido)
	rec = do(t, router, http.MethodPut, "/perfumes/"+itoa(criado.ID), map[string]any{
		"marca_id": marcaID, "nome": "Teste Atualizado", "preco": 250.00,
		"tamanho": "100ml", "genero": "unissex", "estoque": 1,
	}, tk.AccessToken)
	if rec.Code != http.StatusOK || decode[db.Perfume](t, rec).Nome != "Teste Atualizado" {
		t.Fatalf("UPDATE: resposta inesperada %d (%s)", rec.Code, rec.Body.String())
	}

	// DELETE (protegido)
	rec = do(t, router, http.MethodDelete, "/perfumes/"+itoa(criado.ID), nil, tk.AccessToken)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE: esperava 204, veio %d", rec.Code)
	}
	rec = do(t, router, http.MethodGet, "/perfumes/"+itoa(criado.ID), nil, "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("READ após DELETE: esperava 404, veio %d", rec.Code)
	}
}

func TestMarcaComPerfumesAninhados(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	tk := autenticar(t, router, "aninhado@teste.com")
	marcaID := criarMarca(t, router, "Marca Teste Aninhado", tk.AccessToken)
	for _, nome := range []string{"P1", "P2"} {
		rec := do(t, router, http.MethodPost, "/perfumes", map[string]any{
			"marca_id": marcaID, "nome": nome, "preco": 100,
		}, tk.AccessToken)
		if rec.Code != http.StatusCreated {
			t.Fatalf("criar %s: %d (%s)", nome, rec.Code, rec.Body.String())
		}
	}

	rec := do(t, router, http.MethodGet, "/marcas/"+itoa(marcaID), nil, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET marca: esperava 200, veio %d", rec.Code)
	}
	type marcaComPerfumes struct {
		ID       int64        `json:"id"`
		Perfumes []db.Perfume `json:"perfumes"`
	}
	out := decode[marcaComPerfumes](t, rec)
	if out.ID != marcaID || len(out.Perfumes) != 2 {
		t.Fatalf("esperava marca %d com 2 perfumes, veio id=%d com %d", marcaID, out.ID, len(out.Perfumes))
	}
}

func TestPerfumeMarcaInexistente(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	tk := autenticar(t, router, "fk@teste.com")
	rec := do(t, router, http.MethodPost, "/perfumes", map[string]any{
		"marca_id": 999999999, "nome": "Fantasma", "preco": 1,
	}, tk.AccessToken)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400 para marca_id inexistente, veio %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestPerfumeNaoEncontrado(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	rec := do(t, router, http.MethodGet, "/perfumes/999999999", nil, "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404, veio %d", rec.Code)
	}
}

func TestMarcaNomeDuplicado(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	tk := autenticar(t, router, "dupmarca@teste.com")
	criarMarca(t, router, "Marca Duplicada Teste", tk.AccessToken)
	rec := do(t, router, http.MethodPost, "/marcas", map[string]any{"nome": "Marca Duplicada Teste"}, tk.AccessToken)
	if rec.Code != http.StatusConflict {
		t.Fatalf("esperava 409 para nome duplicado, veio %d (%s)", rec.Code, rec.Body.String())
	}
}

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/seunome/perfume-api/internal/auth"
)

type chaveContexto string

const usuarioIDKey chaveContexto = "usuarioID"

// Autenticacao protege rotas exigindo um JWT válido no header Authorization.
// Em caso de sucesso, injeta o ID do usuário no contexto da requisição.
func Autenticacao(svc *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				erroJSON(w, http.StatusUnauthorized, "token de acesso ausente")
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			id, err := svc.ValidarAccessToken(token)
			if err != nil {
				erroJSON(w, http.StatusUnauthorized, "token de acesso inválido ou expirado")
				return
			}

			ctx := context.WithValue(r.Context(), usuarioIDKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UsuarioID recupera o ID do usuário autenticado do contexto.
func UsuarioID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(usuarioIDKey).(int64)
	return id, ok
}

func erroJSON(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"erro":"` + msg + `"}`))
}

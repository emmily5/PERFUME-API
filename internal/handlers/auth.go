package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/seunome/perfume-api/internal/auth"
	"github.com/seunome/perfume-api/internal/db"
)

type credenciais struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type respostaTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiraEmSeg  int    `json:"expira_em_seg"`
}

// Registrar POST /auth/registrar — cria um novo usuário com senha em hash bcrypt.
func (h *Handler) Registrar(w http.ResponseWriter, r *http.Request) {
	var in credenciais
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respErro(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	if !emailValido(in.Email) {
		respErro(w, http.StatusBadRequest, "e-mail inválido")
		return
	}
	if len(in.Senha) < 8 {
		respErro(w, http.StatusBadRequest, "a senha deve ter ao menos 8 caracteres")
		return
	}

	hash, err := auth.HashSenha(in.Senha)
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao processar senha")
		return
	}

	user, err := h.q.CriarUsuario(r.Context(), db.CriarUsuarioParams{Email: in.Email, SenhaHash: hash})
	if violaUnico(err) {
		respErro(w, http.StatusConflict, "e-mail já cadastrado")
		return
	}
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao criar usuário")
		return
	}

	respJSON(w, http.StatusCreated, map[string]any{"id": user.ID, "email": user.Email})
}

// Login POST /auth/login — valida credenciais e emite access + refresh token.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in credenciais
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respErro(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))

	user, err := h.q.BuscarUsuarioPorEmail(r.Context(), in.Email)
	// Mensagem genérica para não revelar se o e-mail existe (enumeração de usuários).
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !auth.VerificarSenha(user.SenhaHash, in.Senha)) {
		respErro(w, http.StatusUnauthorized, "credenciais inválidas")
		return
	}
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao autenticar")
		return
	}

	h.emitirTokens(w, r, user.ID)
}

// Refresh POST /auth/refresh — rotaciona o refresh token: revoga o antigo e emite um novo.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.RefreshToken == "" {
		respErro(w, http.StatusBadRequest, "refresh_token é obrigatório")
		return
	}

	hash := auth.HashToken(in.RefreshToken)
	rt, err := h.q.BuscarRefreshToken(r.Context(), hash)
	if errors.Is(err, pgx.ErrNoRows) {
		respErro(w, http.StatusUnauthorized, "refresh token inválido")
		return
	}
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao validar refresh token")
		return
	}

	// Detecção de reuso: se o token já foi revogado, revoga todos os tokens do usuário.
	if rt.Revogado {
		_ = h.q.RevogarTodosTokensDoUsuario(r.Context(), rt.UserID)
		respErro(w, http.StatusUnauthorized, "refresh token já utilizado")
		return
	}
	if time.Now().After(rt.ExpiraEm.Time) {
		respErro(w, http.StatusUnauthorized, "refresh token expirado")
		return
	}

	// Rotação: revoga o token atual e emite um novo par.
	if err := h.q.RevogarRefreshToken(r.Context(), hash); err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao rotacionar token")
		return
	}
	h.emitirTokens(w, r, rt.UserID)
}

// Logout POST /auth/logout — revoga o refresh token informado.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.RefreshToken == "" {
		respErro(w, http.StatusBadRequest, "refresh_token é obrigatório")
		return
	}
	if err := h.q.RevogarRefreshToken(r.Context(), auth.HashToken(in.RefreshToken)); err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao revogar token")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// emitirTokens gera o access token (JWT) e um novo refresh token persistido (hash).
func (h *Handler) emitirTokens(w http.ResponseWriter, r *http.Request, userID int64) {
	access, err := h.auth.GerarAccessToken(userID)
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao gerar token de acesso")
		return
	}

	raw, hash, err := auth.GerarRefreshToken()
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao gerar refresh token")
		return
	}

	_, err = h.q.CriarRefreshToken(r.Context(), db.CriarRefreshTokenParams{
		UserID:    userID,
		TokenHash: hash,
		ExpiraEm:  pgtype.Timestamptz{Time: time.Now().Add(h.auth.RefreshTTL), Valid: true},
	})
	if err != nil {
		respErro(w, http.StatusInternalServerError, "erro ao salvar refresh token")
		return
	}

	respJSON(w, http.StatusOK, respostaTokens{
		AccessToken:  access,
		RefreshToken: raw,
		TokenType:    "Bearer",
		ExpiraEmSeg:  int(h.auth.AccessTTL.Seconds()),
	})
}

// emailValido faz uma validação simples de formato de e-mail.
func emailValido(email string) bool {
	at := strings.Index(email, "@")
	return at > 0 && at < len(email)-1 && strings.Contains(email[at:], ".")
}

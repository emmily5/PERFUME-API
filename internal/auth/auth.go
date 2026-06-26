// Package auth concentra a lógica de autenticação: JWT de acesso,
// hashing de senha (bcrypt) e geração/hash de refresh tokens.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ErrTokenInvalido é retornado quando um access token não é válido.
var ErrTokenInvalido = errors.New("token inválido")

// Service guarda a configuração de autenticação.
type Service struct {
	secret     []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// NewService cria o serviço de auth. O segredo não pode ser vazio.
func NewService(secret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		secret:     []byte(secret),
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
	}
}

// --- Senhas (bcrypt) ---

// HashSenha gera o hash bcrypt de uma senha em texto puro.
func HashSenha(senha string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	return string(b), err
}

// VerificarSenha compara uma senha em texto puro com o hash armazenado.
func VerificarSenha(hash, senha string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha)) == nil
}

// --- Access token (JWT HS256) ---

// GerarAccessToken cria um JWT assinado com o ID do usuário no claim "sub".
func (s *Service) GerarAccessToken(userID int64) (string, error) {
	agora := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, 10),
		IssuedAt:  jwt.NewNumericDate(agora),
		ExpiresAt: jwt.NewNumericDate(agora.Add(s.AccessTTL)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidarAccessToken valida a assinatura e a expiração, devolvendo o ID do usuário.
func (s *Service) ValidarAccessToken(tokenStr string) (int64, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalido
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return 0, ErrTokenInvalido
	}
	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, ErrTokenInvalido
	}
	return id, nil
}

// --- Refresh token (opaco, aleatório) ---

// GerarRefreshToken cria um token aleatório (valor em claro) e devolve também seu hash.
// Apenas o hash é persistido no banco.
func GerarRefreshToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	return raw, HashToken(raw), nil
}

// HashToken devolve o SHA-256 (hex) de um refresh token.
func HashToken(raw string) string {
	soma := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(soma[:])
}

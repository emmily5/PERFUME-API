package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logger registra no terminal cada requisição recebida: método, URL, status e tempo
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()

		// wrappedWriter captura o status code escrito pelo handler
		ww := &wrappedWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		log.Printf("[%s] %s %s — status %d — %s",
			inicio.Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			ww.statusCode,
			time.Since(inicio),
		)
	})
}

// Recovery evita que um panic derrube o servidor; devolve 500 ao cliente
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC recuperado: %v", err)
				http.Error(w, "erro interno do servidor", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// wrappedWriter é um ResponseWriter que guarda o status code para o logger
type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (ww *wrappedWriter) WriteHeader(code int) {
	ww.statusCode = code
	ww.ResponseWriter.WriteHeader(code)
}

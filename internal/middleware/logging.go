package middleware

import (
	"net/http"
	"time"

	"log/slog"
)

func Logging(baseLog *slog.Logger) func(http.Handler) http.Handler {
	const op = "http.request"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := baseLog.With(slog.String("op", op))

			start := time.Now()
			logger.Info("request started",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote", r.RemoteAddr),
			)

			ww := &responseWriter{ResponseWriter: w}
			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			logger.Info("request finished",
				slog.Int("status", ww.status),
				slog.Duration("duration", duration),
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

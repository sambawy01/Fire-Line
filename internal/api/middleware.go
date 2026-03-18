package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

type ctxKey string

const correlationKey ctxKey = "correlation_id"

func CorrelationIDFrom(ctx context.Context) string {
	v, _ := ctx.Value(correlationKey).(string)
	return v
}

func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cid := r.Header.Get("X-Request-ID")
		if cid == "" {
			b := make([]byte, 16)
			rand.Read(b)
			cid = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", cid)
		ctx := context.WithValue(r.Context(), correlationKey, cid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"correlation_id", CorrelationIDFrom(r.Context()),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "error", err, "correlation_id", CorrelationIDFrom(r.Context()))
				WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

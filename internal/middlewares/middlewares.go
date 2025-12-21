package middlewares

import (
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "userID"
const cookieName = "user_id"
const secretKey = "super-secret-key"

func GzipRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(gr)
			r.Header.Del("Content-Encoding")
		}
		next.ServeHTTP(w, r)
	})
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gw := &responseGzipWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(gw, r)

		if gw.gzipEnabled {
			gw.gzw.Close()
		}
	})
}

type responseGzipWriter struct {
	http.ResponseWriter
	gzw         *gzip.Writer
	gzipEnabled bool
}

func (w *responseGzipWriter) enableGzip() error {
	if w.gzipEnabled {
		return nil
	}
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Del("Content-Length")
	w.gzw = gzip.NewWriter(w.ResponseWriter)
	w.gzipEnabled = true
	return nil
}

func (w *responseGzipWriter) WriteHeader(status int) {
	ct := w.Header().Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "text/html") {
		_ = w.enableGzip()
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseGzipWriter) Write(b []byte) (int, error) {
	ct := w.Header().Get("Content-Type")

	if strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "text/html") {
		if !w.gzipEnabled {
			if err := w.enableGzip(); err != nil {
				return w.ResponseWriter.Write(b)
			}
		}
		return w.gzw.Write(b)
	}

	return w.ResponseWriter.Write(b)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(cookieName)

		if err != nil || !validCookie(c.Value) {
			userID := uuid.NewString()
			value := userID + "|" + sign(userID)

			http.SetCookie(w, &http.Cookie{
				Name:  cookieName,
				Value: value,
				Path:  "/",
			})

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		parts := strings.Split(c.Value, "|")
		if len(parts) != 2 {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, parts[0])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func sign(value string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil))
}

func validCookie(value string) bool {
	parts := strings.Split(value, "|")
	if len(parts) != 2 {
		return false
	}
	return sign(parts[0]) == parts[1]
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserIDKey).(string)
	return id, ok
}

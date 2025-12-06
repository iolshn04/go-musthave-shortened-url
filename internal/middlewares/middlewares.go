package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

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
				// на ошибки gzip.NewWriter редко жалуются; в случае ошибки просто пишем как есть
				return w.ResponseWriter.Write(b)
			}
		}
		return w.gzw.Write(b)
	}

	return w.ResponseWriter.Write(b)
}

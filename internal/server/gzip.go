package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// gzipMiddleware опционально распаковывает тело при Content-Encoding: gzip
// и опционально сжимает ответ при Accept-Encoding: gzip для application/json и text/html.
func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := maybeGunzipRequestBody(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		gzw := newGzipResponseWriter(w, r)
		defer func() { _ = gzw.Close() }()
		next.ServeHTTP(gzw, r)
	})
}

type gzipRequestCloser struct {
	*gzip.Reader
	underlying io.ReadCloser
}

func (c *gzipRequestCloser) Close() error {
	var errR, errU error
	if c.Reader != nil {
		errR = c.Reader.Close()
	}
	if c.underlying != nil {
		errU = c.underlying.Close()
	}
	if errR != nil {
		return errR
	}
	return errU
}

func maybeGunzipRequestBody(r *http.Request) error {
	if r.Body == nil || r.Body == http.NoBody {
		return nil
	}
	if !headerContainsToken(r.Header.Get("Content-Encoding"), "gzip") {
		return nil
	}
	zr, err := gzip.NewReader(r.Body)
	if err != nil {
		return err
	}
	r.Body = &gzipRequestCloser{Reader: zr, underlying: r.Body}
	r.Header.Del("Content-Encoding")
	return nil
}

func headerContainsToken(header, token string) bool {
	token = strings.ToLower(token)
	for _, part := range strings.Split(header, ",") {
		if strings.TrimSpace(strings.ToLower(part)) == token {
			return true
		}
	}
	return false
}

func acceptsGzip(r *http.Request) bool {
	return headerContainsToken(r.Header.Get("Accept-Encoding"), "gzip")
}

func contentTypeBase(ct string) string {
	return strings.TrimSpace(strings.SplitN(ct, ";", 2)[0])
}

func shouldGzipResponse(r *http.Request, ct string) bool {
	if !acceptsGzip(r) {
		return false
	}
	switch contentTypeBase(ct) {
	case "application/json", "text/html":
		return true
	default:
		return false
	}
}

// gzipResponseWriter при первом WriteHeader с подходящим Content-Type включает gzip на тело ответа.
type gzipResponseWriter struct {
	http.ResponseWriter
	r *http.Request

	gz          *gzip.Writer
	wroteHeader bool
}

func newGzipResponseWriter(w http.ResponseWriter, r *http.Request) *gzipResponseWriter {
	return &gzipResponseWriter{ResponseWriter: w, r: r}
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	ct := w.Header().Get("Content-Type")
	if shouldGzipResponse(w.r, ct) {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
		w.gz = gzip.NewWriter(w.ResponseWriter)
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if w.gz != nil {
		return w.gz.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *gzipResponseWriter) Close() error {
	if w.gz != nil {
		return w.gz.Close()
	}
	return nil
}

var _ http.Flusher = (*gzipResponseWriter)(nil)

func (w *gzipResponseWriter) Flush() {
	if w.gz != nil {
		_ = w.gz.Flush()
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

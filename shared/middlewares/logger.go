package middlewares

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// Logger logs HTTP request/response information.
//
// It does not:
//   - handle application errors
//   - recover panics
//   - log request bodies
//   - log response bodies
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			next.ServeHTTP(ww, r)

			duration := time.Since(start)

			attrs := []slog.Attr{
				// Request
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("query", r.URL.RawQuery),
				slog.String("route", routePattern(r)),
				slog.String("protocol", r.Proto),
				slog.String("scheme", requestScheme(r)),
				slog.String("host", r.Host),

				// Client
				slog.String("client_ip", clientIP(r)),
				slog.String("user_agent", r.UserAgent()),
				slog.String("referer", r.Referer()),

				// Request
				slog.Int64("request_content_length", r.ContentLength),

				// Response
				slog.Int("status", ww.status),
				slog.Int64("response_bytes", ww.bytes),

				// Duration
				slog.Duration("duration", duration),
				slog.Int64("duration_ms", duration.Milliseconds()),
			}

			if r.TLS != nil {
				attrs = append(
					attrs,
					slog.String("tls_version", tlsVersion(r.TLS.Version)),
					slog.String("tls_cipher", tlsCipher(r.TLS.CipherSuite)),
				)
			}

			logger.LogAttrs(
				r.Context(),
				statusLogLevel(ww.status),
				"http request",
				attrs...,
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter

	status      int
	bytes       int64
	wroteHeader bool
}

func (w *responseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}

	w.status = status
	w.wroteHeader = true

	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	n, err := w.ResponseWriter.Write(p)
	w.bytes += int64(n)

	return n, err
}

// Flush supports streaming responses.
func (w *responseWriter) Flush() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack supports WebSockets and connection upgrades.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New(
			"http.ResponseWriter does not implement http.Hijacker",
		)
	}

	return h.Hijack()
}

// ReadFrom preserves the optimized io.Copy path.
func (w *responseWriter) ReadFrom(r io.Reader) (int64, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		n, err := rf.ReadFrom(r)
		w.bytes += n

		return n, err
	}

	n, err := io.Copy(w.ResponseWriter, r)
	w.bytes += n

	return n, err
}

// Unwrap exposes the underlying ResponseWriter.
func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func routePattern(r *http.Request) string {
	rc := chi.RouteContext(r.Context())
	if rc == nil {
		return ""
	}

	return rc.RoutePattern()
}

func clientIP(r *http.Request) string {
	// Only trust these headers when they come from a trusted proxy.
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")

		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}

func requestScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}

	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}

	return "http"
}

func statusLogLevel(status int) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

func tlsVersion(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS1.0"
	case tls.VersionTLS11:
		return "TLS1.1"
	case tls.VersionTLS12:
		return "TLS1.2"
	case tls.VersionTLS13:
		return "TLS1.3"
	default:
		return strconv.FormatUint(uint64(version), 16)
	}
}

func tlsCipher(cipher uint16) string {
	return fmt.Sprintf("0x%04x", cipher)
}

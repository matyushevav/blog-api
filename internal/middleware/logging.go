package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// LoggingMiddleware содержит общие middleware: логирование, Recovery и CORS.
type LoggingMiddleware struct {
	logger *log.Logger
}

// NewLoggingMiddleware создает LoggingMiddleware с заданным логгером.
func NewLoggingMiddleware(logger *log.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

// Logger пишет в лог каждый запрос после его обработки.
func (m *LoggingMiddleware) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		m.logger.Printf("%s %s %s %d %s",
			r.RemoteAddr, r.Method, r.URL.Path, rw.statusCode, time.Since(start))
	})
}

// Recovery перехватывает панику в хендлере и отвечает 500 вместо обрыва соединения.
func (m *LoggingMiddleware) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				m.logger.Printf("panic recovered: %v\n%s", rec, debug.Stack())

				respondWithError(w, "internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CORS разрешает запросы к API из браузера с других доменов.
func (m *LoggingMiddleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		next.ServeHTTP(w, r)
	})
}

// responseWriter - обертка над ResponseWriter, чтобы перехватить статус код.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

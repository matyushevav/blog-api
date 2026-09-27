package middleware

import "net/http"

// ToMiddleware преобразует middleware функцию из одного формата в другой.
func ToMiddleware(fn func(http.HandlerFunc) http.HandlerFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return fn(next.ServeHTTP)
	}
}

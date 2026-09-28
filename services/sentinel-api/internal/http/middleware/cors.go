package middleware

import (
	"net/http"
	"strings"
)

func CORS(allowedOrigins []string, next http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = true
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if origin != "" && (allowed[origin] || allowed["*"]) {
			writer.Header().Set("Access-Control-Allow-Origin", origin)
			writer.Header().Set("Vary", "Origin")
			writer.Header().Set("Access-Control-Allow-Credentials", "true")
			writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization")
		}

		if request.Method == http.MethodOptions {
			if origin != "" && !allowed[origin] && !allowed["*"] {
				http.Error(writer, "origen no permitido", http.StatusForbidden)
				return
			}
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		if strings.TrimSpace(origin) != "" && !allowed[origin] && !allowed["*"] {
			http.Error(writer, "origen no permitido", http.StatusForbidden)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

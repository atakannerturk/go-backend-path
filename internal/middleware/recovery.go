package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/atakannerturk/go-backend-path/pkg/logger"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"path", r.URL.Path,
					"method", r.Method,
				)

				respondError(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

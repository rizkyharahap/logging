package middlewares

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := log.Logger.With().
			Str("request_id", uuid.New().String()).
			Str("url", r.URL.String()).
			Str("method", r.Method).
			Logger()
		ctx := log.WithContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

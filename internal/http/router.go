package httpx

import (
	"net/http"

	"gophkeeper/internal/http/handlers"
	"gophkeeper/internal/http/middleware"
	"gophkeeper/internal/logger"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

// NewRouter builds HTTP router.
func NewRouter(jwtSecret string, authH *handlers.AuthHandler, secH *handlers.SecretsHandler) http.Handler {

	r := chi.NewRouter()

	r.Use(logger.RequestLogger)
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)

	// маршруты ---
	r.Route("/api/v1", func(r chi.Router) {

		//  AUTH ----------
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authH.Register)
			r.Post("/login", authH.Login)
		})

		//  PROTECTED -------
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtSecret)) // (JWT) проверка токена
			r.Get("/records/list", secH.List)
			r.Post("/records/upsert", secH.Upsert)
			r.Get("/records/{id}", secH.Get)
			r.Delete("/records/{id}", secH.Delete)
		})
	})

	return r
}

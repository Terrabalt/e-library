package main

import (
	"ic-rhadi/e_library/endpoints"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	zerolog.TimeFieldFormat = time.RFC3339
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.CleanPath)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", endpoints.LoginPostEndpoint())
		r.Post("/google", endpoints.LoginGoogleEndpoint())
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	log.Info().Str("Server port", port).Msg("Server started")
	log.Info().Err(http.ListenAndServe(":"+port, r)).Msg("Server stopped")
}

package main

import (
	"context"
	"fmt"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/endpoints"
	"ic-rhadi/e_library/googlehelper"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type config struct {
	jwtsecret string
	pgHost    string
	pgPort    string
	pgUser    string
	pgPass    string
	pgDB      string
}

func main() {
	log.Logger = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	zerolog.TimeFieldFormat = time.RFC3339

	var jwtSecret string
	var ok bool
	if jwtSecret, ok = os.LookupEnv("JWTSECRET"); !ok {
		log.Info().Msg("Enviroment keys not set up! Loading .env file..")
		err := godotenv.Load()
		if err != nil {
			log.Fatal().Err(err).Caller().Msg("Error loading .env file")
		}
		jwtSecret = os.Getenv("JWTSECRET")
	}

	sessionAuth := jwtauth.New("HS256", []byte(jwtSecret), nil)
	dbUrl := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("PG_HOST"),
		os.Getenv("PG_PORT"),
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASS"),
		os.Getenv("PG_DB"),
	)

	db, err := database.StartDB(dbUrl)
	if err != nil {
		log.Panic().Err(err).Str("database link", dbUrl).Caller().Msg("Error starting database")
	}
	if err := db.InitDB(context.Background()); err != nil {
		log.Panic().Err(err).Caller().Msg("Error initializing database")
	}

	gValidator, err := googlehelper.NewGValidator(context.Background())
	if err != nil {
		log.Panic().Err(err).Msg("Google token validator failed to initialize")
		return
	}

	r := chi.NewRouter()

	r.Use(middleware.CleanPath)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", endpoints.LoginPostEndpoint(db, sessionAuth))
		r.Post("/google", endpoints.LoginGoogleEndpoint(db, sessionAuth, gValidator))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	log.Info().Str("Server port", port).Msg("Server started")
	log.Info().Err(http.ListenAndServe(":"+port, r)).Msg("Server stopped")
}

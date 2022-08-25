package main

import (
	"context"
	"fmt"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/endpoints"
	"ic-rhadi/e_library/googlehelper"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	JWTSecret string `env:"JWTSECRET"`
	PgHost    string `env:"PG_HOST"`
	PgPort    int    `env:"PG_PORT"`
	PgUser    string `env:"PG_USER"`
	PgPass    string `env:"PG_PASS"`
	PgDB      string `env:"PG_DB"`
	Port      int    `env:"PORT,required"`
}

func main() {
	log.Logger = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	zerolog.TimeFieldFormat = time.RFC3339

	if err := godotenv.Load(); err != nil {
		log.Fatal().Err(err).Caller().Msg("Error loading .env file")
	}

	var config Config
	if err := envconfig.Process(context.Background(), &config); err != nil {
		log.Fatal().Err(err).Msg("Required enviroment keys was not set up")
	}

	sessionAuth := jwtauth.New("HS256", []byte(config.JWTSecret), nil)
	dbUrl := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.PgHost,
		config.PgPort,
		config.PgUser,
		config.PgPass,
		config.PgDB,
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

	log.Info().Int("Server port", config.Port).Msg("Server started")
	log.Info().Err(http.ListenAndServe(":"+strconv.Itoa(config.Port), r)).Msg("Server stopped")
}

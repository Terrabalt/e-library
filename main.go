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
	JWTSecret    string       `env:"JWTSECRET"`
	PgHost       string       `env:"PG_HOST"`
	PgPort       int          `env:"PG_PORT"`
	PgUser       string       `env:"PG_USER"`
	PgPass       string       `env:"PG_PASS"`
	PgDB         string       `env:"PG_DB"`
	Port         int          `env:"PORT,required"`
	LoginLengths LoginLengths `env:""`
}

type LoginLengths struct {
	TokenLength   time.Duration `env:"TOKEN_DURATION"`
	SessionLength time.Duration `env:"SESSION_DURATION"`
}

func main() {
	log.Logger = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Caller().Logger()
	zerolog.TimeFieldFormat = time.RFC3339

	if err := godotenv.Load(); err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
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
		log.Panic().Err(err).Str("database link", dbUrl).Msg("Error starting database")
	}
	if err := db.InitDB(context.Background()); err != nil {
		log.Panic().Err(err).Msg("Error initializing database")
	}
	defer db.CloseDB()

	gValidator, err := googlehelper.NewGValidator(context.Background())
	if err != nil {
		log.Panic().Err(err).Msg("Google token validator failed to initialize")
		return
	}

	sessionLength := config.LoginLengths.SessionLength
	tokenLength := config.LoginLengths.TokenLength
	r := chi.NewRouter()

	r.Use(middleware.CleanPath)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", endpoints.LoginPost(db, sessionAuth, sessionLength, tokenLength))
		r.Post("/google", endpoints.LoginGoogle(db, sessionAuth, gValidator, sessionLength, tokenLength))
	})

	log.Info().Int("Server port", config.Port).Msg("Server started")
	log.Debug().Interface("test", config).Send()
	if err = http.ListenAndServe(":"+strconv.Itoa(config.Port), r); err != http.ErrServerClosed {
		log.Error().Err(err).Msg("Server stopped with error")
	} else {
		log.Info().Msg("Server stopped normally")
	}
}

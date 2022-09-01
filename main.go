package main

import (
	"context"
	"fmt"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/emailhelper"
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

type config struct {
	JWTSecret    string       `env:"JWTSECRET"`
	PgHost       string       `env:"PG_HOST"`
	PgPort       int          `env:"PG_PORT"`
	PgUser       string       `env:"PG_USER"`
	PgPass       string       `env:"PG_PASS"`
	PgDB         string       `env:"PG_DB"`
	Port         int          `env:"PORT,required"`
	LoginLengths loginLengths `env:""`
}

type loginLengths struct {
	TokenLength   time.Duration `env:"TOKEN_DURATION"`
	SessionLength time.Duration `env:"SESSION_DURATION"`
}

func main() {
	log.Logger = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Caller().Logger()
	zerolog.TimeFieldFormat = time.RFC3339

	if err := godotenv.Load(); err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}

	var conf config
	if err := envconfig.Process(context.Background(), &conf); err != nil {
		log.Fatal().Err(err).Msg("Required enviroment keys was not set up")
	}

	var email emailhelper.ActivationMailDriver
	email, err := emailhelper.NewActivationMailHelper(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("Required enviroment keys was not set up")
	}

	sessionAuth := jwtauth.New("HS256", []byte(conf.JWTSecret), nil)
	dbURL := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		conf.PgHost,
		conf.PgPort,
		conf.PgUser,
		conf.PgPass,
		conf.PgDB,
	)

	db, err := database.StartDB(dbURL)
	if err != nil {
		log.Panic().Err(err).Str("database link", dbURL).Msg("Error starting database")
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

	sessionLength := conf.LoginLengths.SessionLength
	tokenLength := conf.LoginLengths.TokenLength
	r := chi.NewRouter()

	r.Use(middleware.CleanPath)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", endpoints.LoginPost(db, sessionAuth, sessionLength, tokenLength))
		r.Post("/google", endpoints.LoginGoogle(db, sessionAuth, gValidator, sessionLength, tokenLength))
		r.Post("/register", endpoints.RegisterPost(db, sessionAuth, email))
		r.Post("/register/google", endpoints.RegisterGoogle(db, sessionAuth, gValidator, email))
	})

	log.Info().Int("Server port", conf.Port).Msg("Server started")
	if err = http.ListenAndServe(":"+strconv.Itoa(conf.Port), r); err != http.ErrServerClosed {
		log.Error().Err(err).Msg("Server stopped with error")
	} else {
		log.Info().Msg("Server stopped normally")
	}
}

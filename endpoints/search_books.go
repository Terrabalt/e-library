package endpoints

import (
	"errors"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/sessiontoken"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

var errSearchQueryTooShort = errors.New("search query too short")

func searchBooks(
	db database.BookInterface,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		query := strings.TrimSpace(r.URL.Query().Get("query"))
		if len(query) < 3 {
			render.Render(w, r, BadRequestError(errSearchQueryTooShort))
			return
		}

		_, token, err := jwtauth.FromContext(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Getting book-searcher account failed unexpectedly")
			render.Render(w, r, InternalServerError())
			return
		}

		var sch sessiontoken.TokenClaimsSchema
		if err := sch.FromInterface(token); err != nil {
			log.Error().Err(err).Msg("Getting book-searcher account failed unexpectedly")
			render.Render(w, r, InternalServerError())
			return
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		books, err := db.SearchBooks(ctx, 20, page, query, sch.Email)
		if err != nil {
			log.Error().Err(err).Str("query", query).Str("account", sch.Email).Msg("Searching book failed")
			render.Render(w, r, InternalServerError())
			return
		}
		booksResponse := BooksFromDatabase(books)

		render.Render(w, r, &booksResponse)
	}
}

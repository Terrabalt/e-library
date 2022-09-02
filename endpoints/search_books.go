package endpoints

import (
	"errors"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/sessiontoken"
	"net/http"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

var errSearchQueryMalformed = errors.New("search query not found")
var errSearchQueryTooShort = errors.New("search query too short")

func SearchBooks(
	db database.BookInterface,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := r.Context()

		query := r.URL.Query().Get("query")
		if query == "" {
			log.Debug().Msg("Getting book-search query string returned an error")
			render.Render(w, r, BadRequestError(errSearchQueryMalformed))
			return
		}
		if len(query) < 3 {
			log.Debug().Msg("Getting book-search query string returned an error")
			render.Render(w, r, BadRequestError(errSearchQueryTooShort))
			return
		}

		_, token, err := jwtauth.FromContext(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("Getting book-searcher account failed")
			render.Render(w, r, BadRequestError(ErrSessionTokenMissingOrInvalid))
			return
		}

		var sch sessiontoken.TokenClaimsSchema
		if err := sch.StrictFromInterface(token); err != nil {
			log.Debug().Err(err).Msg("Getting book-searcher account failed")
			render.Render(w, r, BadRequestError(ErrSessionTokenMissingOrInvalid))
			return
		}

		books, err := db.SearchBooks(ctx, query, sch.Email)
		if err != nil {
			log.Error().Err(err).Str("query", query).Str("account", sch.Email).Msg("Searching book failed")
			render.Render(w, r, InternalServerError())
			return
		}
		booksResponse := BooksFromDatabase(books)

		render.Render(w, r, &booksResponse)
	}
}

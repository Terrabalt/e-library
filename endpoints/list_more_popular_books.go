package endpoints

import (
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/sessiontoken"
	"net/http"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

func ListMorePopularBooks(
	db database.BookInterface,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := r.Context()

		_, token, err := jwtauth.FromContext(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("Getting more-popular-books-list asker account failed")
			render.Render(w, r, BadRequestError(ErrSessionTokenMissingOrInvalid))
			return
		}

		var sch sessiontoken.TokenClaimsSchema
		if err := sch.StrictFromInterface(token); err != nil {
			log.Debug().Err(err).Msg("Listing more popular books on Homepage failed")
			render.Render(w, r, BadRequestError(ErrSessionTokenMissingOrInvalid))
			return
		}

		books, err := db.GetPopularBooks(ctx, sch.Email)
		if err != nil {
			log.Error().Err(err).Msg("Listing more popular books on Homepage failed")
			render.Render(w, r, InternalServerError())
			return
		}
		booksResponse := BooksFromDatabase(books)

		render.Render(w, r, &booksResponse)
	}
}

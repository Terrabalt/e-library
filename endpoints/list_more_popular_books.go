package endpoints

import (
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/sessiontoken"
	"net/http"
	"strconv"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

func listMorePopularBooks(
	db database.BookInterface,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		_, token, err := jwtauth.FromContext(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("Getting more-popular-books-list asker account failed")
			render.Render(w, r, InternalServerError())
			return
		}

		var sch sessiontoken.TokenClaimsSchema
		if err := sch.StrictFromInterface(token); err != nil {
			log.Debug().Err(err).Msg("Getting more-popular-books-list asker account failed")
			render.Render(w, r, InternalServerError())
			return
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		books, err := db.GetPopularBooksPaginated(ctx, 20, page, sch.Email)
		if err != nil {
			log.Error().Err(err).Msg("Listing more popular books failed")
			render.Render(w, r, InternalServerError())
			return
		}
		booksResponse := BooksFromDatabase(books)

		render.Render(w, r, &booksResponse)
	}
}

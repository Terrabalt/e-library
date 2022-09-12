package endpoints

import (
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/sessiontoken"
	"net/http"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

func homepageListNewBooks(
	db database.BookInterface,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sch, err := sessiontoken.FromContext(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("Getting new-books-list asker account failed")
			render.Render(w, r, InternalServerError())
			return
		}

		books, err := db.GetNewBooksPaginated(ctx, 8, 0, sch.Email)
		if err != nil {
			log.Error().Err(err).Msg("Listing new books on Homepage failed")
			render.Render(w, r, InternalServerError())
			return
		}
		booksResponse := BooksFromDatabase(books)

		render.Render(w, r, &booksResponse)
	}
}

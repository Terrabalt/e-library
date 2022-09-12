package endpoints

import (
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/sessiontoken"
	"net/http"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

func homepageListPopularBooks(
	db database.BookInterface,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sch, err := sessiontoken.FromContext(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("Getting popular-books-list asker account failed")
			render.Render(w, r, InternalServerError())
			return
		}

		books, err := db.GetPopularBooksPaginated(ctx, 8, 0, sch.Email)
		if err != nil {
			log.Error().Err(err).Msg("Listing popular books on Homepage failed")
			render.Render(w, r, InternalServerError())
			return
		}
		booksResponse := BooksFromDatabase(books)

		render.Render(w, r, &booksResponse)
	}
}

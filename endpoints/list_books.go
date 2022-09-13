package endpoints

import (
	"errors"
	"ic-rhadi/e_library/database"
	"ic-rhadi/e_library/sessiontoken"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

var errListBookCriteriaUnrecognized = errors.New("criteria unrecognized")

func ListBooks(
	db database.BookInterface,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sch, err := sessiontoken.FromContext(ctx)
		if err != nil {
			log.Error().Err(err).Msg("ListBook: Getting account failed unexpectedly")
			render.Render(w, r, InternalServerError())
			return
		}
		if sch == nil {
			log.Error().Msg("ListBook: Getting account returned nil")
			render.Render(w, r, InternalServerError())
			return
		}

		criteria := strings.TrimSpace(r.URL.Query().Get("criteria"))
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 0 {
			render.Render(w, r, InternalServerError())
			return
		}
		log.Debug().Str("criteria", criteria).Int("page", page).Send()

		var books []database.Book
		switch criteria {
		case "newHomepage":
			books, err = db.GetNewBooksPaginated(ctx, 8, 0, sch.Email)
		case "new":
			books, err = db.GetNewBooksPaginated(ctx, 20, page, sch.Email)
		case "popularHomepage":
			books, err = db.GetPopularBooksPaginated(ctx, 8, 0, sch.Email)
		case "popular":
			books, err = db.GetPopularBooksPaginated(ctx, 20, page, sch.Email)
		case "search", "":
			query := strings.TrimSpace(r.URL.Query().Get("query"))
			books, err = db.SearchBooks(ctx, 20, page, query, sch.Email)
		default:
			log.Debug().Str("criteria", criteria).Msg("ListBook: criteria unrecognized")
			render.Render(w, r, BadRequestError(errListBookCriteriaUnrecognized))
			return
		}
		if err != nil {
			log.Error().Err(err).Msg("ListBook: Getting account failed unexpectedly")
			render.Render(w, r, InternalServerError())
			return
		}

		booksResponse := BooksFromDatabase(books)

		render.Render(w, r, &booksResponse)
	}
}

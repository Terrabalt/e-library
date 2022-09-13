package endpoints

import (
	"ic-rhadi/e_library/database"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/render"
)

func ListBooks(
	db database.BookInterface,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//ctx := r.Context()

		criteria := strings.TrimSpace(r.URL.Query().Get("criteria"))
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 0 {
			render.Render(w, r, InternalServerError())
			return
		}
		switch criteria {
		case "newHomepage":
			homepageListNewBooks(db)(w, r)
		case "new":
			listMoreNewBooks(db)(w, r)
		case "popularHomepage":
			homepageListPopularBooks(db)(w, r)
		case "popular":
			listMorePopularBooks(db)(w, r)
		case "search":
		default:
			searchBooks(db)(w, r)
		}
	}
}

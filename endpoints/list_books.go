package endpoints

import (
	"ic-rhadi/e_library/database"
	"net/http"
	"strconv"
	"strings"
)

func ListBooks(
	db database.BookInterface,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//ctx := r.Context()

		criteria := strings.TrimSpace(r.URL.Query().Get("criteria"))
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))

		switch criteria {
		case "new":
			if page > 0 {
				listMoreNewBooks(db)(w, r)
			} else {
				homepageListNewBooks(db)(w, r)
			}
		case "popular":
			if page > 0 {
				listMorePopularBooks(db)(w, r)
			} else {
				homepageListPopularBooks(db)(w, r)
			}
		case "search":
		default:
			searchBooks(db)(w, r)
		}
	}
}

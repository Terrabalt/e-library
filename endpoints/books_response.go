package endpoints

import (
	"ic-rhadi/e_library/database"
	"net/http"

	"github.com/go-chi/render"
)

type BookResponse struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	CoverURL string `json:"cover_url"`
	Summary  string `json:"summary,omitempty"`
	Readers  int    `json:"readers"`
	IsFav    bool   `json:"is_favorite"`
}

func (b *BookResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusOK)
	w.Header().Set("content-type", "application/json")
	return nil
}

func BookFromDatabase(dBook database.Book) BookResponse {
	var b BookResponse

	b.ID = dBook.ID.String()
	b.Title = dBook.Title
	b.Author = dBook.Author
	b.Summary = dBook.Summary
	b.IsFav = dBook.IsFav
	b.CoverURL = dBook.Cover.String()
	b.Readers = dBook.Readers

	return b
}

type BooksResponse struct {
	Data []BookResponse `json:"data"`
}

func (b *BooksResponse) Render(w http.ResponseWriter, r *http.Request) error {
	for i := range b.Data {
		b.Data[i].Summary = ""
	}
	render.Status(r, http.StatusOK)
	w.Header().Set("content-type", "application/json")
	return nil
}

func BooksFromDatabase(dBooks []database.Book) BooksResponse {
	var b BooksResponse

	for _, dBook := range dBooks {
		b.Data = append(b.Data, BookFromDatabase(dBook))
	}

	return b
}

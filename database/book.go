package database

import (
	"context"

	"github.com/google/uuid"
)

type BookInterface interface {
	SearchBooks(ctx context.Context, query string, accountID string) ([]Book, error)
}

var searchBooks = dbStatement{
	nil, `
	SELECT
		b.id,
		b.title,
		b.cover_image,
		b.author,
		b.readers_count,
		EXISTS (
			SELECT
				1
			FROM
				fav_book f
			WHERE
				f.user_id = $1
				AND f.book_id = b.id
		) AS is_favorited
	FROM
		book b
	WHERE
		b.title ILIKE '%' || $2 || '%'
		OR b.author ILIKE '%' || $2 || '%'
	ORDER BY
		b.title ASC;`,
}

func init() {
	prepareStatements = append(prepareStatements,
		&searchBooks,
	)
}

type Book struct {
	ID      uuid.UUID
	Title   string
	Author  string
	Cover   URL
	Summary string
	Readers int
	Rating  float32
	IsFav   bool
}

func (db DBInstance) SearchBooks(ctx context.Context, query string, accountID string) ([]Book, error) {
	var books []Book
	rows, err := searchBooks.Statement.QueryContext(ctx, accountID, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		book := Book{}
		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Cover,
			&book.Author,
			&book.Readers,
			&book.IsFav,
		); err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}

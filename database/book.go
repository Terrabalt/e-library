package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

var getNewBooks = dbStatement{
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
		b.is_new
	ORDER BY
		b.title ASC
	LIMIT
		$2 OFFSET $3;`,
}

type BookInterface interface {
	GetNewBooks(ctx context.Context, limit int, offset int, accountID string) ([]Book, error)
}

func init() {
	prepareStatements = append(prepareStatements,
		&getNewBooks,
	)
}

type Book struct {
	ID      uuid.UUID
	Title   string
	Author  string
	Cover   URL
	Summary string
	Readers int
	IsFav   bool
}

func (db DBInstance) GetNewBooks(ctx context.Context, limit int, offset int, accountID string) ([]Book, error) {
	var books []Book
	lim := sql.NullInt64{Int64: int64(limit), Valid: limit > 0}
	rows, err := getNewBooks.Statement.QueryContext(ctx, accountID, lim, offset)
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

package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type BookInterface interface {
	GetPopularBooks(ctx context.Context, accountID string) ([]Book, error)
	GetPopularBooksPaginated(ctx context.Context, limit int, offset int, accountID string) ([]Book, error)
}

const getPopularBooksStr = `
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
		b.is_popular
	ORDER BY
	b.title ASC%s;`

var getPopularBooks = dbStatement{
	nil, fmt.Sprintf(getPopularBooksStr, ""),
}
var getPopularBooksPaginated = dbStatement{
	nil, fmt.Sprintf(getPopularBooksStr, `
	LIMIT
		$2 OFFSET $3`),
}

func init() {
	prepareStatements = append(prepareStatements,
		&getPopularBooks,
		&getPopularBooksPaginated,
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

func (db DBInstance) GetPopularBooks(ctx context.Context, accountID string) ([]Book, error) {
	var books []Book
	rows, err := getPopularBooks.Statement.QueryContext(ctx, accountID)
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

func (db DBInstance) GetPopularBooksPaginated(ctx context.Context, limit int, offset int, accountID string) ([]Book, error) {
	var books []Book
	if limit <= 0 || offset < 0 {
		return nil, errors.New("function parameters outside the bounds")
	}
	rows, err := getPopularBooks.Statement.QueryContext(ctx, accountID, limit, offset)
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

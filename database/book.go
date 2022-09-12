package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const getNewBooksStr = `
	SELECT
		b.id,
		b.title,
		b.cover_image,
		b.author,
		b.readers_count,
		av.rating,
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
		LEFT JOIN rating_avg AS av ON b.id = av.id 
	WHERE
		b.is_new
	ORDER BY
		b.title ASC%s;`

var getNewBooks = dbStatement{
	nil,
	fmt.Sprintf(getNewBooksStr, ""),
}
var getNewBooksPaginated = dbStatement{
	nil,
	fmt.Sprintf(getNewBooksStr, `
			LIMIT
				$2 OFFSET $3`),
}

const getPopularBooksStr = `
	SELECT
		b.id,
		b.title,
		b.cover_image,
		b.author,
		b.readers_count,
		av.rating,
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
		LEFT JOIN rating_avg AS av ON b.id = av.id 
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

var searchBooks = dbStatement{
	nil, `
	SELECT
		b.id,
		b.title,
		b.cover_image,
		b.author,
		b.readers_count,
		av.rating,
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
		LEFT JOIN rating_avg AS av ON b.id = av.id
	WHERE
		b.title ILIKE '%' || $2 || '%'
		OR b.author ILIKE '%' || $2 || '%'
	ORDER BY
		b.title ASC
	LIMIT
		$3 OFFSET $4;`,
}

type BookInterface interface {
	SearchBooks(ctx context.Context, limit int, offset int, query string, accountID string) ([]Book, error)
	GetNewBooks(ctx context.Context, accountID string) ([]Book, error)
	GetNewBooksPaginated(ctx context.Context, limit int, offset int, accountID string) ([]Book, error)
	GetPopularBooks(ctx context.Context, accountID string) ([]Book, error)
	GetPopularBooksPaginated(ctx context.Context, limit int, offset int, accountID string) ([]Book, error)
}

func init() {
	prepareStatements = append(prepareStatements,
		&getNewBooks,
		&getNewBooksPaginated,
		&getPopularBooks,
		&getPopularBooksPaginated,
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

func (db DBInstance) GetNewBooks(ctx context.Context, accountID string) ([]Book, error) {
	var books []Book
	rows, err := getNewBooks.Statement.QueryContext(ctx, accountID)
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
			&book.Rating,
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

func (db DBInstance) GetNewBooksPaginated(ctx context.Context, limit int, offset int, accountID string) ([]Book, error) {
	var books []Book
	if limit <= 0 || offset < 0 {
		return nil, errors.New("function parameters outside the bounds")
	}
	rows, err := getNewBooksPaginated.Statement.QueryContext(ctx, accountID, limit, offset)
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
			&book.Rating,
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
			&book.Rating,
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
	rows, err := getPopularBooksPaginated.Statement.QueryContext(ctx, accountID, limit, offset)
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
			&book.Rating,
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

func (db DBInstance) SearchBooks(ctx context.Context, limit int, offset int, query string, accountID string) ([]Book, error) {
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
			&book.Rating,
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

package database

import (
	"github.com/google/uuid"
)

type BookInterface interface {
}

func init() {
	prepareStatements = append(prepareStatements)
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

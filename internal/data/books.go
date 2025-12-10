package data

import (
	"database/sql"
	"time"
)

type Book struct {
	ID              int
	Title           string
	Author          string
	PublishDate     time.Time
	ISBN            string
	Description     string
	CopiesTotal     int
	CopiesAvailable int
	Version         int
}

type BookModel struct {
	DB *sql.DB
}

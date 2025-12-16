package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

type Book struct {
	ID              int
	Title           string
	Author          string
	PublishDate     time.Time
	ISBN            string
	Description     string
	CoverImage      string
	Genres          []string
	Pages           int
	Language        string
	Publisher       string
	CopiesTotal     int
	CopiesAvailable int
	Version         int
}

type BookModel struct {
	DB *sql.DB
}

func (m BookModel) GetBooks(limit, offset int) ([]*Book, error) {
	query := `
		SELECT id, title, author, publish_date, isbn, description, cover_image, genres, copies_total, copies_available
		FROM books
		ORDER BY copies_available DESC
		LIMIT $1 OFFSET $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*Book
	for rows.Next() {
		var b Book
		var genres []string
		if err := rows.Scan(
			&b.ID, &b.Title, &b.Author, &b.PublishDate, &b.ISBN, &b.Description, &b.CoverImage,
			pq.Array(&genres), &b.CopiesTotal, &b.CopiesAvailable,
		); err != nil {
			return nil, err
		}
		b.Genres = genres
		books = append(books, &b)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}

func (m BookModel) GetBookByID(id int) (*Book, error) {
	query := `
		SELECT id, title, author, publish_date, isbn, description, cover_image, genres, pages, language, publisher, copies_total, copies_available, version
		FROM books
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var b Book
	var genres []string

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&b.ID,
		&b.Title,
		&b.Author,
		&b.PublishDate,
		&b.ISBN,
		&b.Description,
		&b.CoverImage,
		pq.Array(&genres),
		&b.Pages,
		&b.Language,
		&b.Publisher,
		&b.CopiesTotal,
		&b.CopiesAvailable,
		&b.Version,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	b.Genres = genres
	return &b, nil
}

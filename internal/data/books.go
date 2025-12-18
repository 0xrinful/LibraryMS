package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

var (
	ErrAlreadyBorrowed   = errors.New("book already borrowed by user")
	ErrNoAvailableCopies = errors.New("no available copies")
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

func (m BookModel) BorrowBook(userID, bookID int64, days int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var available int
	err = tx.QueryRowContext(ctx,
		`SELECT copies_available FROM books WHERE id = $1`,
		bookID,
	).Scan(&available)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrRecordNotFound
		}
		return err
	}

	if available < 1 {
		return ErrNoAvailableCopies
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO borrow_records (user_id, book_id, due_at)
		VALUES ($1, $2, NOW() + ($3 || ' days')::interval)
	`, userID, bookID, days)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrAlreadyBorrowed
		}
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE books
		SET copies_available = copies_available - 1,
		    version = version + 1
		WHERE id = $1
	`, bookID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m BookModel) ReturnBook(userID, bookID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE borrow_records
		SET returned_at = NOW()
		WHERE user_id = $1
		  AND book_id = $2
		  AND returned_at IS NULL
	`, userID, bookID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrRecordNotFound
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE books
		SET copies_available = copies_available + 1,
		    version = version + 1
		WHERE id = $1
	`, bookID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m BookModel) Search(q, category, availability, sort string) ([]*Book, error) {
	sqlStr := `
        SELECT id, title, author, publish_date, isbn, description, cover_image, genres, pages,
               language, publisher, copies_total, copies_available, version
        FROM books
        WHERE (title ILIKE $1 OR author ILIKE $1 OR isbn ILIKE $1)
    `
	args := []any{"%" + q + "%"}

	if category != "" && category != "All Categories" {
		sqlStr += " AND $2 = ANY(genres)"
		args = append(args, category)
	}

	if availability != "" {
		switch availability {
		case "Available":
			sqlStr += " AND copies_available > 0"
		case "Borrowed":
			sqlStr += " AND copies_available = 0"
		}
	}

	switch sort {
	case "Title (A-Z)":
		sqlStr += " ORDER BY title ASC"
	case "Title (Z-A)":
		sqlStr += " ORDER BY title DESC"
	case "Author (A-Z)":
		sqlStr += " ORDER BY author ASC"
	case "Newest First":
		sqlStr += " ORDER BY publish_date DESC"
	default:
		sqlStr += " ORDER BY title ASC"
	}

	rows, err := m.DB.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*Book
	for rows.Next() {
		var b Book
		var genres []string
		if err := rows.Scan(
			&b.ID, &b.Title, &b.Author, &b.PublishDate, &b.ISBN, &b.Description,
			&b.CoverImage, pq.Array(&genres), &b.Pages, &b.Language, &b.Publisher,
			&b.CopiesTotal, &b.CopiesAvailable, &b.Version,
		); err != nil {
			return nil, err
		}
		b.Genres = genres
		books = append(books, &b)
	}
	return books, nil
}

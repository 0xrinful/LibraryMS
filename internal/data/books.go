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
	ErrDuplicateISBN     = errors.New("duplicate ISBN")
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

func (m BookModel) ISBNExists(isbn string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM books WHERE isbn = $1)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var exists bool
	err := m.DB.QueryRowContext(ctx, query, isbn).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (m BookModel) ISBNExistsExcluding(isbn string, excludeID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM books WHERE isbn = $1 AND id != $2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var exists bool
	err := m.DB.QueryRowContext(ctx, query, isbn, excludeID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
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

func (m BookModel) Count() (int, error) {
	query := `SELECT COUNT(*) FROM books`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	err := m.DB.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m BookModel) GetAll() ([]*Book, error) {
	query := `
		SELECT id, title, author, publish_date, isbn, description, cover_image, genres, pages, language, publisher, copies_total, copies_available, version
		FROM books
		ORDER BY title ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}

func (m BookModel) Insert(book *Book) error {
	query := `
		INSERT INTO books (title, author, publish_date, isbn, description, cover_image, genres, pages, language, publisher, copies_total, copies_available)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		book.Title,
		book.Author,
		book.PublishDate,
		book.ISBN,
		book.Description,
		book.CoverImage,
		pq.Array(book.Genres),
		book.Pages,
		book.Language,
		book.Publisher,
		book.CopiesTotal,
		book.CopiesAvailable,
	}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&book.ID, &book.Version)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrDuplicateISBN
		}
		return err
	}
	return nil
}

func (m BookModel) Update(book *Book) error {
	query := `
		UPDATE books
		SET title = $1, author = $2, publish_date = $3, isbn = $4, description = $5, 
		    cover_image = $6, genres = $7, pages = $8, language = $9, publisher = $10, 
		    copies_total = $11, copies_available = $12, version = version + 1
		WHERE id = $13
		RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		book.Title,
		book.Author,
		book.PublishDate,
		book.ISBN,
		book.Description,
		book.CoverImage,
		pq.Array(book.Genres),
		book.Pages,
		book.Language,
		book.Publisher,
		book.CopiesTotal,
		book.CopiesAvailable,
		book.ID,
	}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&book.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				return ErrDuplicateISBN
			}
			return err
		}
	}
	return nil
}

func (m BookModel) Delete(id int) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM books WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

package data

import (
	"database/sql"
	"time"
)

type BorrowedBook struct {
	BorrowID   int64
	BorrowedAt time.Time
	ReturnedAt *time.Time
	DueAt      time.Time

	BookID     int
	Title      string
	Author     string
	CoverImage string
}

type BorrowRecordModel struct {
	DB *sql.DB
}

func (m BorrowRecordModel) GetCurrentBorrows(userID int64) ([]*BorrowedBook, error) {
	query := `
		SELECT b.id, b.title, b.author, b.cover_image, br.borrowed_at, br.due_at
		FROM borrow_records br
		INNER JOIN books b ON br.book_id = b.id
		WHERE br.user_id = $1 AND br.returned_at IS NULL
		ORDER BY br.borrowed_at DESC
	`
	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var borrows []*BorrowedBook
	for rows.Next() {
		var bb BorrowedBook
		if err := rows.Scan(&bb.BookID, &bb.Title, &bb.Author, &bb.CoverImage, &bb.BorrowedAt, &bb.DueAt); err != nil {
			return nil, err
		}
		borrows = append(borrows, &bb)
	}
	return borrows, nil
}

func (m BorrowRecordModel) GetBorrowHistory(userID int64) ([]*BorrowedBook, error) {
	query := `
		SELECT b.id, b.title, b.author, b.cover_image, br.borrowed_at, br.due_at, br.returned_at
		FROM borrow_records br
		INNER JOIN books b ON br.book_id = b.id
		WHERE br.user_id = $1 AND br.returned_at IS NOT NULL
		ORDER BY br.returned_at DESC
	`
	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*BorrowedBook
	for rows.Next() {
		var bb BorrowedBook
		if err := rows.Scan(&bb.BookID, &bb.Title, &bb.Author, &bb.CoverImage, &bb.BorrowedAt, &bb.DueAt, &bb.ReturnedAt); err != nil {
			return nil, err
		}
		history = append(history, &bb)
	}
	return history, nil
}

func (m BorrowRecordModel) CountActiveBorrows() (int, error) {
	query := `SELECT COUNT(*) FROM borrow_records WHERE returned_at IS NULL`

	var count int
	err := m.DB.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m BorrowRecordModel) CountOverdue() (int, error) {
	query := `SELECT COUNT(*) FROM borrow_records WHERE returned_at IS NULL AND due_at < NOW()`

	var count int
	err := m.DB.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

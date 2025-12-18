package data

import (
	"database/sql"
	"time"
)

type BorrowRecord struct {
	ID         int64
	UserID     int64
	BookID     int64
	BorrowedAt time.Time
	ReturnedAt *time.Time
	DueAt      *time.Time
}

type BorrowedBook struct {
	BorrowID   int64
	BorrowedAt time.Time
	DueAt      time.Time

	BookID     int
	Title      string
	Author     string
	CoverImage string
}

type BorrowRecordModel struct {
	DB *sql.DB
}

func (m BorrowRecordModel) CurrentByUser(userID int64) ([]*BorrowedBook, error) {
	query := `
		SELECT
			br.id,
			br.borrowed_at,
			br.due_at,

			b.id,
			b.title,
			b.author,
			b.cover_image
		FROM borrow_records br
		INNER JOIN books b ON b.id = br.book_id
		WHERE br.user_id = $1
		  AND br.returned_at IS NULL
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

		err := rows.Scan(
			&bb.BorrowID,
			&bb.BorrowedAt,
			&bb.DueAt,
			&bb.BookID,
			&bb.Title,
			&bb.Author,
			&bb.CoverImage,
		)
		if err != nil {
			return nil, err
		}

		borrows = append(borrows, &bb)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return borrows, nil
}

func (m BorrowRecordModel) ActiveCountByUser(userID int64) (int, error) {
	var count int

	err := m.DB.QueryRow(`
		SELECT COUNT(*)
		FROM borrow_records
		WHERE user_id = $1
		  AND returned_at IS NULL
	`, userID).Scan(&count)

	return count, err
}

func (m BorrowRecordModel) TotalCountByUser(userID int64) (int, error) {
	var count int

	err := m.DB.QueryRow(`
		SELECT COUNT(*)
		FROM borrow_records
		WHERE user_id = $1
	`, userID).Scan(&count)

	return count, err
}

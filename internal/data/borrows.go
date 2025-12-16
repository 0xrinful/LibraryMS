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
}

type BorrowRecordModel struct {
	DB *sql.DB
}

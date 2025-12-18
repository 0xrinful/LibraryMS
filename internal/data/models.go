package data

import (
	"database/sql"
	"errors"
)

var ErrRecordNotFound = errors.New("models: record not found")

type Models struct {
	Users interface {
		Insert(user *User) error
		GetByEmail(email string) (*User, error)
		Get(id int64) (*User, error)
	}

	Books interface {
		GetBooks(limit, offset int) ([]*Book, error)
		GetBookByID(id int) (*Book, error)
		BorrowBook(userID, bookID int64, days int) error
		ReturnBook(userID, bookID int64) error
	}

	BorrowRecord interface {
		GetCurrentBorrows(userID int64) ([]*BorrowedBook, error)
		GetBorrowHistory(userID int64) ([]*BorrowedBook, error)
	}
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users:        UserModel{DB: db},
		Books:        BookModel{DB: db},
		BorrowRecord: BorrowRecordModel{DB: db},
	}
}

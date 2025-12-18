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
		BorrowBook(userID, bookID int64) error
	}
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users: UserModel{DB: db},
		Books: BookModel{DB: db},
	}
}

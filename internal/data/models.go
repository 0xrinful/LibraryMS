package data

import "database/sql"

type Models struct {
	Users interface{}
	Books interface{}
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users: UserModel{DB: db},
		Books: BookModel{DB: db},
	}
}

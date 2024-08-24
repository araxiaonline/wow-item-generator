package sqlite

import (
	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"
)

type SqlLite struct {
	*sqlx.DB
}

var SqlLiteDb *SqlLite

func Connect(path string) (*SqlLite, error) {
	client, err := sqlx.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	SqlLiteDb = &SqlLite{client}
	return SqlLiteDb, nil
}

func GetDb() (*SqlLite, error) {
	return SqlLiteDb, nil
}

func (db *SqlLite) Close() {
	if db.DB != nil {
		db.DB.Close()
	}
}

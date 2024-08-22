package sqlite

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type SqlLite struct {
	client *sql.DB
}

func Connect(path string) (*SqlLite, error) {
	client, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	return &SqlLite{client: client}, nil
}

func (db *SqlLite) Close() {
	db.client.Close()
}

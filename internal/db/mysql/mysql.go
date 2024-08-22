package db

import (
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type MySqlDb struct {
	*sqlx.DB
}

type MySqlConfig struct {
	Host     string
	User     string
	Password string
	Database string
}

var MySql MySqlDb

func ConnectMySql(config *MySqlConfig) (*MySqlDb, error) {

	if config == nil {
		config = &MySqlConfig{
			Host:     os.Getenv("DB_HOST"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Database: os.Getenv("DB_NAME"),
		}
	}

	connString := config.User + ":" + config.Password + "@tcp(" + config.Host + ")/" + config.Database
	client, err := sqlx.Open("mysql", connString)
	if err != nil {
		return nil, err
	}

	Conn = &MySql{client}
	return Conn, nil
}

func (db *MySql) Close() {
	db.Close()
}

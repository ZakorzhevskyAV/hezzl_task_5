package postgresql

import (
	"database/sql"
	"fmt"
	"os"
)

var DBConn *sql.DB

func DBConnect() error {
	connString := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"))

	conn, err := sql.Open("postgres", connString)
	if err != nil {
		return err
	}

	DBConn = conn

	return nil
}

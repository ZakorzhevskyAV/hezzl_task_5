package postgresql

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

var DBConn *sql.DB

func DBConnect() error {

	connString := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_DB"))

	conn, err := sql.Open("postgres", connString)
	if err != nil {
		return err
	}
	err = conn.Ping()
	if err != nil {
		return err
	}

	DBConn = conn

	return nil
}

package clickhouse_logging

import (
	"database/sql"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"os"
)

var ClickhouseDBConn *sql.DB

func CHConnect() error {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", os.Getenv("CLICKHOUSE_HOST"), 8123)},
		Auth: clickhouse.Auth{
			Database: os.Getenv("CLICKHOUSE_DB"),
			Username: os.Getenv("CLICKHOUSE_USER"),
			Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		},
		Protocol: clickhouse.HTTP,
	})
	ClickhouseDBConn = conn
	err := conn.Ping()
	if err != nil {
		return err
	}
	return nil
}

package db

import (
	"log"

	"github.com/ClickHouse/clickhouse-go/v2"
)

var (
	Connection clickhouse.Conn
)

func init() {
	db := "localhost:9000"

	var err error
	Connection, err = clickhouse.Open(&clickhouse.Options{
		Addr: []string{db},
		Auth: clickhouse.Auth{
			Database: "stratometrics",
			// todo: auth
		},
	})
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
}

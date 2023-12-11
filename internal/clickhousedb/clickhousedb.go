package clickhousedb

import (
	"fmt"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"

	"github.com/vexxhost/stratometrics/internal/clickhousedb/migrations"
)

type Database struct {
	Connection clickhouse.Conn
}

func Open() (*Database, error) {
	if err := godotenv.Load(); err != nil {
		log.WithError(err).Warn("could not load .env file")
	}

	dsn := os.Getenv("CLICKHOUSE_DSN")
	db := "default"

	if err := migrations.Up(fmt.Sprintf("clickhouse://%s/%s", dsn, db)); err != nil {
		return nil, err
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{dsn},
		Auth: clickhouse.Auth{
			Database: db,
			// todo: auth
		},
	})
	if err != nil {
		return nil, err
	}

	return &Database{
		Connection: conn,
	}, nil
}

func (db *Database) Close() error {
	return db.Connection.Close()
}

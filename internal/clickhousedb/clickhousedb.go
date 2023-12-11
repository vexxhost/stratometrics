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

	host := os.Getenv("CLICKHOUSE_HOST")
	username := os.Getenv("CLICKHOUSE_USERNAME")
	password := os.Getenv("CLICKHOUSE_PASSWORD")
	db := "default"

	if err := migrations.Up(fmt.Sprintf("clickhouse://%s:%s@%s/%s?x-cluster-name=stratometrics&x-migrations-table-engine=ReplicatedMergeTree", username, password, host, db)); err != nil {
		return nil, err
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{host},
		Auth: clickhouse.Auth{
			Username: username,
			Password: password,
			Database: db,
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

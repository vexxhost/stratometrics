package main

import (
	"log"

	"github.com/vexxhost/stratometrics/internal/clickhousedb"
	"github.com/vexxhost/stratometrics/internal/router"
)

func main() {
	db, err := clickhousedb.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := router.NewRouter(db)
	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"fmt"
	"log"

	"github.com/vexxhost/stratometrics/internal/router"
	"github.com/vexxhost/stratometrics/migrations"
)

func main() {
	db := "localhost:9000"

	err := migrations.Up(
		fmt.Sprintf("clickhouse://%s", db),
	)
	if err != nil {
		panic(err)
	}

	r := router.NewRouter()

	// Run the server
	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}

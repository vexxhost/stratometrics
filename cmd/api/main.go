package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/vexxhost/stratometrics/internal/database"
	"github.com/vexxhost/stratometrics/internal/router"
)

func main() {
	db, err := database.Open()
	if err != nil {
		log.Fatal(err)
	}

	r := router.NewRouter(db)
	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}

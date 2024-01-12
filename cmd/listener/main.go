package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/wagslane/go-rabbitmq"

	"github.com/vexxhost/stratometrics/internal/consumers"
	"github.com/vexxhost/stratometrics/internal/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.WithError(err).Warn("could not load .env file")
	}

	db, err := database.Open()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := rabbitmq.NewConn(
		os.Getenv("NOVA_TRANSPORT_URL"),
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	nova, err := consumers.NewNovaConsumer(db, conn)
	if err != nil {
		log.Fatal(err)
	}
	defer nova.Close()

	// block main thread - wait for shutdown signal
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	// TODO: thread 2
	// check for all non-deleted instances in mysql
	// ensure they are still equiv status in nova
	// if not, update mysql

	fmt.Println("awaiting signal")
	<-done
	fmt.Println("stopping consumer")
}

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/wagslane/go-rabbitmq"

	"github.com/vexxhost/stratometrics/internal/clickhousedb"
	"github.com/vexxhost/stratometrics/internal/consumers"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	db, err := clickhousedb.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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

	fmt.Println("awaiting signal")
	<-done
	fmt.Println("stopping consumer")

	// thread 2
	// check for all non-deleted instances in clickhouse
	// ensure they are still equiv status in nova
	// if not, update clickhouse

	// thread 3
	// i would say check all in nova but we get periodic nova instance exists
	// health check to make sure we are still getting periodic events
}

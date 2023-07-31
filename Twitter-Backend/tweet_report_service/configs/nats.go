package configs

import (
	"github.com/nats-io/nats.go"
	"log"
	"tweet_report_service/startup/config"
)

func ConnectToNats(config *config.Config) *nats.Conn {

	conn, err := nats.Connect(config.NatsURI)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

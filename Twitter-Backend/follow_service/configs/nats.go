package configs

import (
	"follow_service/startup/config"
	"github.com/nats-io/nats.go"
	"log"
)

func ConnectToNats(config *config.Config) *nats.Conn {

	conn, err := nats.Connect(config.NatsURI)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

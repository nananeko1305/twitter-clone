package main

import (
	"tweet_service/startup"
	cfg "tweet_service/startup/config"
)

func main() {
	config := cfg.NewConfig()
	server := startup.NewServer(config)
	server.Start()
}

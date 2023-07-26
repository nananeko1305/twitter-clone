package main

import (
	"tweet_report_service/startup"
	"tweet_report_service/startup/config"
)

func main() {

	cfg := config.NewConfig()
	server := startup.NewServer(cfg)
	server.Start()
}

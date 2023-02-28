package main

import (
	"follow_service/startup"
	"follow_service/startup/config"
)

func main() {
	cfg := config.NewConfig()
	server := startup.NewServer(cfg)
	server.Start()

}

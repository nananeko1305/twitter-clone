package main

import (
	"auth_service/startup"
	"auth_service/startup/config"
)

func main() {
	cfg := config.NewConfig()
	server := startup.NewServer(cfg)
	//mailerSetup()
	server.Start()
}

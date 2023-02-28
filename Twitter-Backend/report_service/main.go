package main

import (
	"report_service/startup"
	"report_service/startup/config"
)

func main() {
	cfg := config.NewConfig()
	server := startup.NewServer(cfg)
	//mailerSetup()
	server.Start()
}

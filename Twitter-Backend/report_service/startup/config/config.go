package config

import "os"

type Config struct {
	Port string

	EventDB                    string
	NatsHost                   string
	NatsPort                   string
	NatsUser                   string
	NatsPass                   string
	JaegerAddress              string
	ReportDBHost               string
	ReportDBPort               string
	CreateReportCommandSubject string
	CreateReportReplySubject   string
}

func NewConfig() *Config {
	return &Config{
		Port:                       os.Getenv("REPORT_SERVICE_PORT"),
		EventDB:                    os.Getenv("EVENT_DB"),
		NatsHost:                   os.Getenv("NATS_HOST"),
		NatsPort:                   os.Getenv("NATS_PORT"),
		NatsUser:                   os.Getenv("NATS_USER"),
		NatsPass:                   os.Getenv("NATS_PASS"),
		JaegerAddress:              os.Getenv("JAEGER_ADDRESS"),
		ReportDBHost:               os.Getenv("REPORT_DB_HOST"),
		ReportDBPort:               os.Getenv("REPORT_DB_PORT"),
		CreateReportCommandSubject: os.Getenv("CREATE_REPORT_COMMAND_SUBJECT"),
		CreateReportReplySubject:   os.Getenv("CREATE_REPORT_REPLY_SUBJECT"),
	}
}

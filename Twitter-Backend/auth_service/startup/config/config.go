package config

import "os"

type Config struct {
	Port                     string
	AuthDBHost               string
	AuthDBPort               string
	AuthCacheHost            string
	AuthCachePort            string
	NatsHost                 string
	NatsPort                 string
	NatsUser                 string
	NatsPass                 string
	CreateUserCommandSubject string
	CreateUserReplySubject   string
	JaegerAddress            string
}

func NewConfig() *Config {
	return &Config{
		Port:                     os.Getenv("AUTH_SERVICE_PORT"),
		AuthDBHost:               os.Getenv("AUTH_DB_HOST"),
		AuthDBPort:               os.Getenv("AUTH_DB_PORT"),
		AuthCacheHost:            os.Getenv("AUTH_CACHE_HOST"),
		AuthCachePort:            os.Getenv("AUTH_CACHE_PORT"),
		NatsHost:                 os.Getenv("NATS_HOST"),
		NatsPort:                 os.Getenv("NATS_PORT"),
		NatsUser:                 os.Getenv("NATS_USER"),
		NatsPass:                 os.Getenv("NATS_PASS"),
		CreateUserCommandSubject: os.Getenv("CREATE_USER_COMMAND_SUBJECT"),
		CreateUserReplySubject:   os.Getenv("CREATE_USER_REPLY_SUBJECT"),
		JaegerAddress:            os.Getenv("JAEGER_ADDRESS"),
	}
}

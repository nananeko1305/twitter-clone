package config

import "os"

type Config struct {
	Port                     string
	FollowDBHost             string
	FollowDBPort             string
	FollowDBUser             string
	FollowDBPass             string
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
		Port:                     os.Getenv("FOLLOW_SERVICE_PORT"),
		FollowDBHost:             os.Getenv("FOLLOW_DB_HOST"),
		FollowDBPort:             os.Getenv("FOLLOW_DB_PORT"),
		FollowDBUser:             os.Getenv("FOLLOW_DB_USER"),
		FollowDBPass:             os.Getenv("FOLLOW_DB_PASS"),
		NatsHost:                 os.Getenv("NATS_HOST"),
		NatsPort:                 os.Getenv("NATS_PORT"),
		NatsUser:                 os.Getenv("NATS_USER"),
		NatsPass:                 os.Getenv("NATS_PASS"),
		CreateUserCommandSubject: os.Getenv("CREATE_USER_COMMAND_SUBJECT"),
		CreateUserReplySubject:   os.Getenv("CREATE_USER_REPLY_SUBJECT"),
		JaegerAddress:            os.Getenv("JAEGER_ADDRESS"),
	}
}

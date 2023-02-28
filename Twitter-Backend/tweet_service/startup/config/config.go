package config

import "os"

type Config struct {
	Port                       string
	TweetDB                    string
	NatsHost                   string
	NatsPort                   string
	NatsUser                   string
	NatsPass                   string
	JaegerAddress              string
	TweetCacheHost             string
	TweetCachePort             string
	CreateReportCommandSubject string
	CreateReportReplySubject   string
}

func NewConfig() *Config {
	return &Config{
		Port:                       os.Getenv("TWEET_SERVICE_PORT"),
		TweetDB:                    os.Getenv("TWEET_DB"),
		NatsHost:                   os.Getenv("NATS_HOST"),
		NatsPort:                   os.Getenv("NATS_PORT"),
		NatsUser:                   os.Getenv("NATS_USER"),
		NatsPass:                   os.Getenv("NATS_PASS"),
		JaegerAddress:              os.Getenv("JAEGER_ADDRESS"),
		TweetCacheHost:             os.Getenv("TWEET_CACHE_HOST"),
		TweetCachePort:             os.Getenv("TWEET_CACHE_PORT"),
		CreateReportCommandSubject: os.Getenv("CREATE_REPORT_COMMAND_SUBJECT"),
		CreateReportReplySubject:   os.Getenv("CREATE_REPORT_REPLY_SUBJECT"),
	}
}

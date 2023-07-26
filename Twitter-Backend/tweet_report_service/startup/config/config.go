package config

import "os"

type Config struct {
	ServicePort string
	ServiceHost string
}

func NewConfig() *Config {
	return &Config{
		ServiceHost: os.Getenv("TWEET_REPORT_SERVICE_HOST"),
		ServicePort: os.Getenv("TWEET_REPORT_SERVICE_PORT"),
	}
}

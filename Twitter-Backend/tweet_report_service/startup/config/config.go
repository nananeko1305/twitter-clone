package config

import "os"

type Config struct {
	ServicePort    string
	ServiceHost    string
	DBPort         string
	DBHost         string
	NatsURI        string
	ElasticAddress string
}

func NewConfig() *Config {
	return &Config{
		ServiceHost:    os.Getenv("TWEET_REPORT_SERVICE_HOST"),
		ServicePort:    os.Getenv("TWEET_REPORT_SERVICE_PORT"),
		DBPort:         os.Getenv("TWEET_REPORT_DB_PORT"),
		DBHost:         os.Getenv("TWEET_REPORT_DB_HOST"),
		NatsURI:        os.Getenv("NATS_URI"),
		ElasticAddress: os.Getenv("ELASTICSEARCH_HOSTS"),
	}
}

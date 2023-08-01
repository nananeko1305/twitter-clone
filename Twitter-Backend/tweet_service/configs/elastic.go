package configs

import (
	"github.com/elastic/go-elasticsearch/v7"
	"tweet_service/startup/config"
)

func ConnectToElastic(config *config.Config) (*elasticsearch.Client, error) {

	cfg := elasticsearch.Config{
		Addresses: []string{
			config.ELASTICSEARCH_HOSTS,
		},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

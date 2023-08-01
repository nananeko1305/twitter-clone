package configs

import (
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"tweet_service/startup/config"
)

func ConnectToElastic(config *config.Config) (*esapi.API, error) {

	cfg := elasticsearch.Config{
		Addresses: []string{
			config.ELASTICSEARCH_HOSTS,
		},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	api := esapi.New(client)

	return api, nil
}

package configs

import (
	"github.com/elastic/go-elasticsearch/v7"
	"log"
	"tweet_service/startup/config"
)

func ConnectToElastic(config *config.Config) (*elasticsearch.Client, error) {

	log.Println(config.ELASTICSEARCH_HOSTS)

	cfg := elasticsearch.Config{
		Addresses: []string{
			config.ELASTICSEARCH_HOSTS,
		},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	info, err := client.Info()
	if err != nil {
		return nil, err
	}
	log.Println(info)

	return client, nil
}

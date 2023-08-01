package configs

import (
	"github.com/elastic/go-elasticsearch/v8"
	"log"
	"tweet_report_service/startup/config"
)

func ConnectToElastic(config *config.Config) *elasticsearch.Client {

	cfg := elasticsearch.Config{
		Addresses: []string{
			config.ElasticAddress,
		},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Elasticsearch connection error: %s", err)
	}

	return client
}

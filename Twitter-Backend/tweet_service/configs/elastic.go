package configs

import (
	"crypto/tls"
	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8"
	"log"
	"net"
	"net/http"
	"time"
	"tweet_service/startup/config"
)

func ConnectToElastic(config *config.Config) (*elasticsearch.Client, error) {

	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
		Username: "foo",
		Password: "bar",
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}
	es, _ := elasticsearch.NewClient(cfg)
	log.Print(es.Transport.(*elastictransport.Client).URLs())

	//cfg := elasticsearch.Config{
	//	Addresses: []string{
	//		config.ELASTICSEARCH_HOSTS,
	//	},
	//}
	//
	//client, err := elasticsearch.NewClient(cfg)
	//if err != nil {
	//	return nil, fmt.Errorf("elasticsearch connection error: %w", err)
	//}

	return es, nil
}

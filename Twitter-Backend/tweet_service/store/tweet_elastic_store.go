package store

import (
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"log"
	"strings"
	"tweet_service/domain"
)

type TweetElasticStoreImpl struct {
	client     *elasticsearch.Client
	elasticApi *esapi.API
}

var (
	indexName = "tweets"
)

func NewTweetElasticStoreImpl(client *elasticsearch.Client, elasticApi *esapi.API) domain.TweetElasticStore {
	return &TweetElasticStoreImpl{
		client:     client,
		elasticApi: elasticApi,
	}
}

func (repository *TweetElasticStoreImpl) Get(id string) error {

	return nil

}

func (repository *TweetElasticStoreImpl) Post(tweet domain.Tweet) error {

	jsonData, _ := json.Marshal(tweet)
	jsonString := string(jsonData)

	request := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: tweet.ID.String(),
		Body:       strings.NewReader(jsonString),
		Refresh:    "true",
	}

	response, err := request.Do(context.Background(), repository.client)
	if err != nil {
		return err
	}

	if response.StatusCode == 200 {
		log.Println("Tweet is inserted!")
	}

	return nil

}

func (repository *TweetElasticStoreImpl) Put(tweet domain.Tweet) error {
	return nil

}

func (repository *TweetElasticStoreImpl) Delete(id string) error {
	return nil
}

func (repository *TweetElasticStoreImpl) CheckIndex() error {

	var index []string
	index = append(index, "tweets")

	exists, err := repository.elasticApi.Indices.Exists(index)
	if err != nil {
		return err
	}

	if exists.StatusCode == 404 {
		_, err := repository.elasticApi.Indices.Create(indexName)
		if err != nil {
			return err
		}
	}
	return nil
}

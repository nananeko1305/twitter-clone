package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/olivere/elastic/v7"
	"io/ioutil"
	"log"
	"strings"
	"tweet_service/domain"
)

type TweetElasticStoreImpl struct {
	client         *elasticsearch.Client
	elasticApi     *esapi.API
	olivereElastic *elastic.Client
}

var (
	indexName = "tweets"
)

func NewTweetElasticStoreImpl(client *elasticsearch.Client, elasticApi *esapi.API, oliverElastic *elastic.Client) domain.TweetElasticStore {
	return &TweetElasticStoreImpl{
		client:         client,
		elasticApi:     elasticApi,
		olivereElastic: oliverElastic,
	}
}

func (repository *TweetElasticStoreImpl) Get(id string) error {

	response, err := repository.elasticApi.Get(indexName, id, func(request *esapi.GetRequest) {
		request.Do(context.Background(), repository.client)
	})
	if err != nil {
		return err
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	//get data from response
	var data map[string]interface{}

	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return err
	}

	// Extract the "_source" map from the given data
	sourceMap, ok := data["_source"].(map[string]interface{})
	if !ok {
		log.Println("Failed to extract _source map")
		return err
	}

	// Convert the "_source" map into JSON bytes
	sourceJSON, err := json.Marshal(sourceMap)
	if err != nil {
		log.Println("Error converting _source to JSON:", err)
		return err
	}

	// Unmarshal the JSON bytes into the Tweet struct
	var tweet domain.Tweet
	if err := json.Unmarshal(sourceJSON, &tweet); err != nil {
		log.Println("Error unmarshaling JSON to Tweet struct:", err)
		return err
	}

	return nil

}

func (repository *TweetElasticStoreImpl) GetAll() ([]*domain.Tweet, error) {

	// Searching with olivere/elastic client
	response, err := repository.olivereElastic.Search().
		Index(indexName).
		Query(elastic.NewMatchAllQuery()).
		Do(context.Background())
	if err != nil {
		log.Fatalf("Error executing the search: %s", err)
	}

	// Mapping every tweet from json to domain.Tweet
	var tweets []domain.Tweet
	for _, hit := range response.Hits.Hits {
		var tweet domain.Tweet
		err := json.Unmarshal(hit.Source, &tweet)
		if err != nil {
			log.Printf("Error unmarshaling tweet: %s", err)
			continue
		}
		tweets = append(tweets, tweet)
	}

	log.Println(tweets)

	return nil, nil

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

	if response.StatusCode == 201 {
		log.Println("Tweet is inserted!")
	}

	return nil

}

func (repository *TweetElasticStoreImpl) Put(tweet *domain.Tweet) error {

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

	log.Println(response.StatusCode)

	if response.StatusCode == 200 {
		log.Println("Tweet is updated!")
	}

	return nil

}

func (repository *TweetElasticStoreImpl) Delete(id string) error {

	response, err := repository.client.Delete(indexName, id, func(request *esapi.DeleteRequest) {
		_, err := request.Do(context.Background(), repository.client)
		if err != nil {
			return
		}

	})
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New("error in deleting tweet")
	}

	return nil
}

func (repository *TweetElasticStoreImpl) CheckIndex() error {
	exists, err := repository.elasticApi.Indices.Exists([]string{indexName})
	if err != nil {
		return err
	}

	if exists.StatusCode == 404 {
		settings := `{
			"settings": {
				"analysis": {
					"analyzer": {
						"hashtag_analyzer": {
							"type": "custom",
							"tokenizer": "standard",
							"filter": ["lowercase", "preserve_hashtags"]
						}
					},
					"filter": {
						"preserve_hashtags": {
							"type": "pattern_replace",
							"pattern": "(#\\w+)",
							"replacement": "$1"
						}
					}
				}
			},
			"mappings": {
				"properties": {
					"text": {
						"type": "text",
						"analyzer": "hashtag_analyzer"
					}
				}
			}
		}`

		_, err := repository.elasticApi.Indices.Create(indexName, repository.elasticApi.Indices.Create.WithBody(strings.NewReader(settings)))
		if err != nil {
			return err
		}
	}
	return nil
}

func (repository *TweetElasticStoreImpl) Search(search domain.Search) ([]*domain.Tweet, error) {

	var tweets []*domain.Tweet

	var query elastic.Query

	if search.SearchType == "fuzzy" {
		query = elastic.NewFuzzyQuery(search.Field, search.SearchSTR).Fuzziness("2")
	} else if search.SearchType == "match" {
		query = elastic.NewMatchQuery(search.Field, search.SearchSTR)
	} else {
		log.Fatal("Invalid search type specified")
	}

	// Build the search request with the bool query
	searchResult, err := repository.olivereElastic.Search().
		Index("tweets"). // Index to search in
		Query(query).
		Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error executing the search: %w", err)
	}

	// Process the search results
	for _, hit := range searchResult.Hits.Hits {
		var tweet domain.Tweet
		err := json.Unmarshal(hit.Source, &tweet)
		if err != nil {
			log.Printf("error unmarshaling tweet: %s", err)
			continue
		}
		tweets = append(tweets, &tweet)
	}

	log.Println(tweets)

	return tweets, nil

}

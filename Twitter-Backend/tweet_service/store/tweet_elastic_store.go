package store

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/olivere/elastic/v7"
	"io/ioutil"
	"log"
	"strings"
	"tweet_service/domain"
)

type TweetElasticStoreImpl struct {
	client        *elasticsearch.Client
	elasticApi    *esapi.API
	oliverElastic *elastic.Client
}

var (
	indexName = "tweets"
)

func NewTweetElasticStoreImpl(client *elasticsearch.Client, elasticApi *esapi.API, oliverElastic *elastic.Client) domain.TweetElasticStore {
	return &TweetElasticStoreImpl{
		client:        client,
		elasticApi:    elasticApi,
		oliverElastic: oliverElastic,
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

	//query := `{"query": {"match_all": {}}}`
	//
	//response, err := repository.client.Search(
	//	repository.client.Search.WithContext(context.Background()),
	//	repository.client.Search.WithIndex(indexName),
	//	repository.client.Search.WithBody(bytes.NewReader([]byte(query))),
	//)
	//if err != nil {
	//	log.Fatalf("Error sending the search request: %s", err)
	//}
	//
	//defer response.Body.Close()
	//
	//if response.IsError() {
	//	log.Fatalf("Error response: %s", response.Status())
	//}
	//
	//var data map[string]interface{}
	//if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
	//	log.Fatalf("Error parsing the response: %s", err)
	//}
	//
	//log.Println(data["hits"])

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

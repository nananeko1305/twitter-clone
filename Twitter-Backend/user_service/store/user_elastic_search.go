package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
	"log"
	"user_service/domain"
)

type UserElasticStoreImpl struct {
	olivereElastic *elastic.Client
}

const (
	indexName = "users"
)

func NewUserElasticStoreImpl(olivereElastic *elastic.Client) domain.UserElasticStore {
	return &UserElasticStoreImpl{
		olivereElastic: olivereElastic,
	}
}

func (store UserElasticStoreImpl) Get(username string) (*domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func (store UserElasticStoreImpl) Post(user domain.User) error {

	jsonUser, err := json.Marshal(&user)
	if err != nil {
		log.Println("Error in marshaling user to JSON")
		return err
	}

	response, err := store.olivereElastic.Index().
		Index(indexName).
		BodyString(string(jsonUser)).
		Do(context.Background())
	if err != nil {
		return err
	}

	if response.Status == 201 {
		log.Println("User saved to elastic")
	}

	return nil
}

func (store UserElasticStoreImpl) CheckIndex() {

	exists, err := store.olivereElastic.IndexExists(indexName).Do(context.Background())
	if err != nil {
		log.Fatalf("Error checking index existence: %s", err)
	}

	if !exists {
		// Create the index if it doesn't exist
		createIndex, err := store.olivereElastic.CreateIndex(indexName).Do(context.Background())
		if err != nil {
			log.Fatalf("Error creating index: %s", err)
		}
		if !createIndex.Acknowledged {
			log.Fatalf("Index creation not acknowledged")
		}
		log.Printf("Index '%s' created!\n", indexName)
	}

}

func (store UserElasticStoreImpl) Search(search domain.Search) ([]*domain.User, error) {

	// Initialize the slice to store the search results
	var users []*domain.User

	// Create a bool query to combine multiple field-based match queries
	boolQuery := elastic.NewBoolQuery()

	// Loop through each search string and field, and create a match query for each with fuzziness
	for i, searchStr := range search.SearchSTRs {
		if i < len(search.Fields) {
			matchQuery := elastic.NewMatchQuery(search.Fields[i], searchStr).
				Fuzziness("2") // Set fuzziness to 2 characters
			boolQuery = boolQuery.Must(matchQuery)
		}
	}

	// Build the search request with the bool query
	searchResult, err := store.olivereElastic.Search().
		Index(indexName).
		Query(boolQuery).
		Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error executing the search: %w", err)
	}

	// Process the search results
	for _, hit := range searchResult.Hits.Hits {
		var user domain.User
		err := json.Unmarshal(hit.Source, &user)
		if err != nil {
			log.Printf("error unmarshaling user: %s", err)
			continue
		}
		users = append(users, &user)
	}

	return users, nil

}

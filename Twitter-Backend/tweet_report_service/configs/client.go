package configs

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func ConnectToMongoDB(host, port string) *mongo.Client {

	uri := fmt.Sprintf("mongodb://%s:%s", host, port)
	opt := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(context.Background(), opt)
	if err != nil {
		log.Println("Error in creating client for Mongodb: ", err)
		return nil
	} else {
		log.Println("Successfully connected to MongoDB")
	}

	return client
}

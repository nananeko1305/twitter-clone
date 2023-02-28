package store

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"log"
)

func GetRedisClient(host, port string) (*redis.Client, error) {
	address := fmt.Sprintf("%s:%s", host, port)
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		logrus.Errorln(err)
		log.Printf("failed to ping cache db because of: %s ", err)
		return nil, err
	}

	return client, nil
}

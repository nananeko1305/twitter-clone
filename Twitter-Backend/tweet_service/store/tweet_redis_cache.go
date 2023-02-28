package store

import (
	"context"
	"github.com/go-redis/redis"
	"go.opentelemetry.io/otel/trace"
	"log"
	"time"
	"tweet_service/domain"
)

type TweetRedisCache struct {
	client *redis.Client
	tracer trace.Tracer
}

func NewTweetRedisCache(client *redis.Client, tracer trace.Tracer) domain.TweetCache {
	return &TweetRedisCache{
		client: client,
		tracer: tracer,
	}
}

func (cache *TweetRedisCache) PostCacheData(ctx context.Context, key string, value *[]byte) error {
	ctx, span := cache.tracer.Start(ctx, "TweetRedisCache.PostCacheData")
	defer span.End()

	log.Println("redis post")
	result := cache.client.Set(key, *value, 10*time.Minute)
	log.Println(result.Err())
	log.Println(result.Result())
	if result.Err() != nil {
		log.Printf("redis set error: %s", result.Err())
		return result.Err()
	}

	return nil
}

func (cache *TweetRedisCache) GetCachedValue(ctx context.Context, key string) (*[]byte, error) {
	ctx, span := cache.tracer.Start(ctx, "TweetRedisCache.GetCachedValue")
	defer span.End()

	result := cache.client.Get(key)

	if result.Err() == nil {
		token, err := result.Bytes()
		if err != nil {
			return nil, err
		}
		return &token, nil
	}
	return nil, nil
}

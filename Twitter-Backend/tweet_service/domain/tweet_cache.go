package domain

import "context"

type TweetCache interface {
	PostCacheData(ctx context.Context, key string, value *[]byte) error
	GetCachedValue(ctx context.Context, key string) (*[]byte, error)
}

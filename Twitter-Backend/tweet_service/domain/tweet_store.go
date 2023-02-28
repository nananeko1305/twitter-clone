package domain

import (
	"context"
	"github.com/gocql/gocql"
)

type TweetStore interface {
	GetPostsFeedByUser(ctx context.Context, usernames []string) ([]*Tweet, error)
	GetRecommendAdsForUser(ctx context.Context, ids []string) ([]*Tweet, error)
	//SaveImageRedis(imageBytes []byte) error
	GetAll(ctx context.Context) ([]Tweet, error)
	GetTweetsByUser(ctx context.Context, username string) ([]*Tweet, error)
	Post(ctx context.Context, tweet *Tweet) (*Tweet, error)
	Favorite(ctx context.Context, id string, username string) (int, error)
	GetLikesByTweet(ctx context.Context, tweetID string) ([]*Favorite, error)
	Retweet(ctx context.Context, tweetID string, username string) (*gocql.UUID, int, error)
	SaveImage(ctx context.Context, tweetID gocql.UUID, imageBytes []byte) error
	GetTweetImage(ctx context.Context, id string) ([]byte, error)
	GetOne(ctx context.Context, tweetID string) (*Tweet, error)
}

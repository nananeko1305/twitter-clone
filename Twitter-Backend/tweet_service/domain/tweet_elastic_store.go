package domain

type TweetElasticStore interface {
	Get(id string) error
	Post(tweet Tweet) error
	Put(tweet Tweet) error
	Delete(id string) error
	CheckIndex() error
}

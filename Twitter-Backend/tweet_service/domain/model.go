package domain

import (
	"github.com/gocql/gocql"
)

type Tweet struct {
	ID            gocql.UUID `json:"id"`
	Text          string     `json:"text"`
	CreatedAt     int64      `json:"created_on"`
	Favorited     bool       `json:"favorited"`
	FavoriteCount int        `json:"favorite_count"`
	Retweeted     bool       `json:"retweeted"`
	RetweetCount  int        `json:"retweet_count"`
	Username      string     `json:"username"`
	OwnerUsername string     `json:"owner_username"`
	Image         bool       `json:"image"`
	Advertisement bool       `json:"advertisement"`
}

type AdConfig struct {
	TweetID   string `json:"tweet_id"`
	Residence string `json:"residence"`
	Gender    string `json:"gender"`
	AgeFrom   int    `json:"age_from"`
	AgeTo     int    `json:"age_to"`
}

type AdTweet struct {
	Tweet    Tweet    `json:"tweet"`
	AdConfig AdConfig `json:"config"`
}

type Favorite struct {
	TweetID  gocql.UUID `json:"tweet_id"`
	Username string     `json:"username"`
	ID       gocql.UUID `json:"id"`
}

type Retweet struct {
	TweetID  gocql.UUID `json:"tweet_id"`
	Username string     `json:"username"`
	ID       gocql.UUID `json:"id"`
}

type TweetID struct {
	ID string `json:"id"`
}

type Timespent struct {
	TweetID   string `json:"tweet_id"`
	Timespent int64  `json:"timespent"`
}

type Event struct {
	TweetID   string
	Type      string
	Timestamp int
}

type FeedInfo struct {
	Usernames []string `json:"usernames"`
	AdIds     []string `json:"ad_ids"`
}

type FeedData struct {
	Feed []*Tweet `json:"feed"`
	Ads  []*Tweet `json:"ads"`
}

//id, ad , created at, favcount, favorited, image, owusername, retweetcount, retweeted, text, username
// bf4054c7-7ea8-4e5e-9714-03b8b880e28f | True | 1675109581 | 1 | True | True |  | 1 | True | Reklama | Company123
//INSERT INTO tweet(id, advertisement, created_at, favorite)

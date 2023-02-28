package store

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"tweet_service/domain"
	"tweet_service/errors"
	"github.com/gocql/gocql"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

const (
	DATABASE               = "tweet"
	COLLECTION             = "tweet"
	COLLECTION_BY_USER     = "tweets_by_user"
	COLLECTION_FAVORITE    = "favorite"
	COLLECTION_RETWEET     = "retweet"
	COLLECTION_TWEET_IMAGE = "tweet_image"
)

type TweetRepo struct {
	session *gocql.Session
	logger  *log.Logger
	tracer  trace.Tracer
	logging *logrus.Logger
}

func New(logger *log.Logger, tracer trace.Tracer, logging *logrus.Logger) (*TweetRepo, error) {
	db := os.Getenv("TWEET_DB")
	log.Println(db)

	cluster := gocql.NewCluster(db)
	cluster.Keyspace = "system"
	session, err := cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	err = session.Query(
		fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
					WITH replication = {
						'class' : 'SimpleStrategy',
						'replication_factor' : %d
					}`, DATABASE, 1)).Exec()
	if err != nil {
		logger.Println(err)
	}
	session.Close()

	cluster.Keyspace = DATABASE
	cluster.Consistency = gocql.One
	session, err = cluster.CreateSession()

	if err != nil {
		logger.Println(err)
		return nil, err
	}

	return &TweetRepo{
		session: session,
		logger:  logger,
		tracer:  tracer,
		logging: logging,
	}, nil
}

func (sr *TweetRepo) CloseSession() {
	sr.session.Close()
}

// Field picture is missing
func (sr *TweetRepo) CreateTables() {
	err := sr.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
					(id UUID, text text, created_at time, favorited boolean, favorite_count int, retweeted boolean,
					retweet_count int, username text, owner_username text, image boolean, advertisement boolean,
					PRIMARY KEY ((id), created_at))
					WITH CLUSTERING ORDER BY (created_at DESC)`, //for now there is no clustering order!!
			COLLECTION)).Exec()

	err = sr.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
					(id UUID, text text, created_at time, favorited boolean, favorite_count int, retweeted boolean,
					retweet_count int, username text, owner_username text, image boolean, advertisement boolean,
					PRIMARY KEY ((username), created_at))
					WITH CLUSTERING ORDER BY (created_at DESC)`, //clustering key by creating date and pk for tweet id and user_id
			COLLECTION_BY_USER)).Exec()

	err = sr.session.Query(
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id UUID, tweet_id UUID, username text, PRIMARY KEY ((tweet_id), username))",
			COLLECTION_FAVORITE)).Exec()

	err = sr.session.Query(
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (tweet_id UUID, image blob, PRIMARY KEY ((tweet_id)))",
			COLLECTION_TWEET_IMAGE)).Exec()

	err = sr.session.Query(
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id UUID, tweet_id UUID, username text, PRIMARY KEY ((tweet_id), username))",
			COLLECTION_RETWEET)).Exec()

	if err != nil {
		sr.logger.Printf("CASSANDRA CREATE TABLE ERR: %s", err.Error())
	}
}

// insert into tweet (tweet_id, created_at, favorite_count, favorited, retweet_count, retweeted, text, user_id) values
// (60089906-68d2-11ed-9022-0242ac120002, 1641540002, 0, false, 0, false, 'cao', dae71a94-68d2-11ed-9022-0242ac120002) ;
func (sr *TweetRepo) GetAll(ctx context.Context) ([]domain.Tweet, error) {
	scanner := sr.session.Query(`SELECT * FROM tweet`).Iter().Scanner()
	sr.logging.Infoln("Store: getAll reached")
	var tweets []domain.Tweet
	for scanner.Next() {
		var tweet domain.Tweet

		err := scanner.Scan(&tweet.ID, &tweet.CreatedAt, &tweet.Advertisement, &tweet.FavoriteCount, &tweet.Favorited,
			&tweet.Image, &tweet.OwnerUsername, &tweet.RetweetCount, &tweet.Retweeted, &tweet.Text, &tweet.Username)
		if err != nil {
			sr.logging.Errorln(err)
			return nil, err
		}

		tweets = append(tweets, tweet)
	}

	if err := scanner.Err(); err != nil {
		sr.logging.Errorln(err)
		return nil, err
	}
	return tweets, nil
}

func (sr *TweetRepo) GetOne(ctx context.Context, tweetID string) (*domain.Tweet, error) {
	ctx, span := sr.tracer.Start(ctx, "TweetStore.GetOne")
	defer span.End()

	sr.logging.Infoln("Store: getOne reached")

	scanner := sr.session.Query(`SELECT * FROM tweet WHERE id = ?`, tweetID).Iter().Scanner()

	var tweets []domain.Tweet
	for scanner.Next() {
		var tweet domain.Tweet

		err := scanner.Scan(&tweet.ID, &tweet.CreatedAt, &tweet.Advertisement, &tweet.FavoriteCount, &tweet.Favorited,
			&tweet.Image, &tweet.OwnerUsername, &tweet.RetweetCount, &tweet.Retweeted, &tweet.Text, &tweet.Username)
		if err != nil {
			sr.logging.Errorln(err)
			return nil, err
		}

		tweets = append(tweets, tweet)
	}

	if err := scanner.Err(); err != nil {
		sr.logging.Errorln(err)
		return nil, err
	}
	return &tweets[0], nil
}

func (sr *TweetRepo) GetTweetsByUser(ctx context.Context, username string) ([]*domain.Tweet, error) {
	ctx, span := sr.tracer.Start(ctx, "TweetStore.GetTweetsByUser")
	defer span.End()
	sr.logging.Infoln("Store: tweetsByUser reached")

	scanner := sr.session.Query(`SELECT * FROM tweets_by_user WHERE username = ?`, username).Iter().Scanner()

	var tweets []*domain.Tweet
	for scanner.Next() {
		var tweet domain.Tweet
		err := scanner.Scan(&tweet.Username, &tweet.CreatedAt, &tweet.Advertisement, &tweet.FavoriteCount,
			&tweet.Favorited, &tweet.ID, &tweet.Image, &tweet.OwnerUsername, &tweet.RetweetCount, &tweet.Retweeted, &tweet.Text)
		if err != nil {
			sr.logging.Errorln(err)
			return nil, err
		}

		tweets = append(tweets, &tweet)
	}

	if err := scanner.Err(); err != nil {
		sr.logging.Errorln(err)
		return nil, err
	}
	return tweets, nil
}

func (sr *TweetRepo) GetPostsFeedByUser(ctx context.Context, usernames []string) ([]*domain.Tweet, error) {
	ctx, span := sr.tracer.Start(ctx, "TweetStore.GetPostsFeedByUser")
	defer span.End()

	query := sr.session.Query(`SELECT * FROM tweets_by_user WHERE username IN ? ORDER BY created_at DESC`, usernames)

	sr.logging.Infoln("Store: getFeed reached")


	query.PageSize(0)
	scanner := query.Iter().Scanner()

	var tweets []*domain.Tweet
	for scanner.Next() {
		var tweet domain.Tweet
		err := scanner.Scan(&tweet.Username, &tweet.CreatedAt, &tweet.Advertisement, &tweet.FavoriteCount,
			&tweet.Favorited, &tweet.ID, &tweet.Image, &tweet.OwnerUsername, &tweet.RetweetCount, &tweet.Retweeted, &tweet.Text)

		if err != nil {
			sr.logging.Errorln(err)
			return nil, err
		}

		tweets = append(tweets, &tweet)
	}

	if err := scanner.Err(); err != nil {
		sr.logging.Errorln(err)
		return nil, err
	}
	return tweets, nil
}

func (sr *TweetRepo) GetRecommendAdsForUser(ctx context.Context, ids []string) ([]*domain.Tweet, error) {
	ctx, span := sr.tracer.Start(ctx, "TweetStore.GetRecommendAdsForUser")
	defer span.End()
	log.Printf("ads: %s", ids)
	query := sr.session.Query(`SELECT * FROM tweet WHERE id IN ? ORDER BY created_at DESC`, ids)
	query.PageSize(0)
	scanner := query.Iter().Scanner()

	var ads []*domain.Tweet
	for scanner.Next() {
		var tweet domain.Tweet
		err := scanner.Scan(&tweet.ID, &tweet.CreatedAt, &tweet.Advertisement, &tweet.FavoriteCount, &tweet.Favorited,
			&tweet.Image, &tweet.OwnerUsername, &tweet.RetweetCount, &tweet.Retweeted, &tweet.Text, &tweet.Username)

		if err != nil {
			sr.logger.Println(err)
			return nil, err
		}

		ads = append(ads, &tweet)
	}

	if err := scanner.Err(); err != nil {
		sr.logger.Println(err)
		return nil, err
	}
	return ads, nil
}

func (sr *TweetRepo) Post(ctx context.Context, tweet *domain.Tweet) (*domain.Tweet, error) {
	ctx, span := sr.tracer.Start(ctx, "TweetStore.Post")
	defer span.End()

	sr.logging.Infoln("Store: post reached")

	insert := fmt.Sprintf("INSERT INTO %s "+
		"(id, created_at, favorite_count, favorited, retweet_count, retweeted, text, username, owner_username, image, advertisement) "+
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", COLLECTION)

	insertByUser := fmt.Sprintf("INSERT INTO %s "+
		"(id, created_at, favorite_count, favorited, retweet_count, retweeted, text, username, owner_username, image, advertisement) "+
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", COLLECTION_BY_USER)

	err := sr.session.Query(
		insert, tweet.ID, tweet.CreatedAt, tweet.FavoriteCount, tweet.Favorited,
		tweet.RetweetCount, tweet.Retweeted, tweet.Text, tweet.Username, tweet.OwnerUsername, tweet.Image, tweet.Advertisement).Exec()

	err = sr.session.Query(
		insertByUser, tweet.ID, tweet.CreatedAt, tweet.FavoriteCount, tweet.Favorited,
		tweet.RetweetCount, tweet.Retweeted, tweet.Text, tweet.Username, tweet.OwnerUsername, tweet.Image, tweet.Advertisement).Exec()

	if err != nil {
		sr.logging.Errorln(err)
		return nil, err
	}
	return tweet, nil
}

func (sr *TweetRepo) SaveImage(ctx context.Context, tweetID gocql.UUID, imageBytes []byte) error {
	ctx, span := sr.tracer.Start(ctx, "TweetStore.SaveImage")
	defer span.End()

	sr.logging.Infoln("Store: saveImage reached")

	insert := fmt.Sprintf("INSERT INTO %s (tweet_id, image) VALUES (?, ?)", COLLECTION_TWEET_IMAGE)

	err := sr.session.Query(insert, tweetID, imageBytes).Exec()
	if err != nil {
		sr.logging.Errorln(err)
		return err
	}

	return nil
}

func (sr *TweetRepo) Favorite(ctx context.Context, tweetID string, username string) (int, error) {
	ctx, span := sr.tracer.Start(ctx, "TweetStore.Favorite")
	defer span.End()

	sr.logging.Infoln("Store: favorite reached")

	id, err := gocql.ParseUUID(tweetID)
	if err != nil {
		sr.logging.Errorln(err)
		return -1, nil
	}

	scanner := sr.session.Query(`SELECT * FROM favorite WHERE tweet_id = ? AND username = ?`, id.String(), username).Iter().Scanner()

	var favorites []*domain.Favorite

	for scanner.Next() {
		var favorite domain.Favorite
		err = scanner.Scan(&favorite.TweetID, &favorite.Username, &favorite.ID)
		if err != nil {
			sr.logging.Errorln(err)
			log.Printf("Error int TweetCassandraStore, Favorite(): %s", err.Error())
			return 502, err
		}

		favorites = append(favorites, &favorite)
	}

	scanner = sr.session.Query(`SELECT * FROM tweet WHERE id = ?`, id.String()).Iter().Scanner()

	sr.logging.Infoln("Getting all tweets in favorite")
	var tweetUsername string
	var tweets []*domain.Tweet
	for scanner.Next() {
		var tweet domain.Tweet
		err := scanner.Scan(&tweet.ID, &tweet.CreatedAt, &tweet.Advertisement, &tweet.FavoriteCount, &tweet.Favorited,
			&tweet.Image, &tweet.OwnerUsername, &tweet.RetweetCount, &tweet.Retweeted, &tweet.Text, &tweet.Username)
		tweetUsername = tweet.Username
		if err != nil {
			sr.logging.Errorln(err)
			log.Printf("Error int TweetCassandraStore, Favorite(): %s", err.Error())
			return 500, err
		}

		tweets = append(tweets, &tweet)
	}

	if len(tweets) == 0 {
		sr.logging.Errorln("no such tweet")
		log.Printf("Error int TweetCassandraStore, Favorite(): %s", err.Error())
		return 500, nil
	}

	favorited := false
	create := false
	favoriteCount := 0
	createdAt := tweets[0].CreatedAt

	if len(favorites) != 0 {
		favoriteCount = tweets[0].FavoriteCount - 1
		favorited = false
		create = false
	} else {
		favoriteCount = tweets[0].FavoriteCount + 1
		favorited = true
		create = true
	}

	isDeleted := false

	if create {
		insert := fmt.Sprintf("INSERT INTO %s "+"(id, tweet_id, username) "+"VALUES (?, ?, ?)", COLLECTION_FAVORITE)
		idFav, _ := gocql.RandomUUID()

		err = sr.session.Query(
			insert, idFav.String(), tweets[0].ID.String(), username).Exec()
		if err != nil {
			log.Printf("Error int TweetCassandraStore, Favorite(): %s", err.Error())
			sr.logger.Println(err)
			return 502, err
		}

	} else {
		deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE tweet_id = ? AND username = ?", COLLECTION_FAVORITE)

		err = sr.session.Query(deleteQuery, id, username).Exec()

		if err != nil {
			sr.logging.Errorf(err.Error())
			return 502, err
		}

		isDeleted = true
	}

	err = sr.session.Query(
		`UPDATE tweet SET favorited=?, favorite_count=? where id=? and created_at = ?`, favorited, favoriteCount, id.String(), createdAt).Exec()

	if err != nil {
		sr.logging.Errorln(err)
		return 502, err
	}

	err = sr.session.Query(
		`UPDATE tweets_by_user SET favorited=?, favorite_count=? where username=? and created_at=?`,
		favorited, favoriteCount, tweetUsername, createdAt).Exec()

	if err != nil {
		sr.logging.Errorln(err)
		return 502, err
	}

	if isDeleted {
		return 200, nil
	}

	return 201, nil
}

func (sr *TweetRepo) GetLikesByTweet(ctx context.Context, tweetID string) ([]*domain.Favorite, error) {
	ctx, span := sr.tracer.Start(ctx, "TweetStore.GetLikesByTweet")
	defer span.End()

	sr.logging.Infoln("Store: getLikesByTweet reached")

	id, err := gocql.ParseUUID(tweetID)
	if err != nil {
		sr.logging.Errorln(err)
		return nil, err
	}

	scanner := sr.session.Query(`SELECT * FROM favorite WHERE tweet_id = ?`, id.String()).Iter().Scanner()

	var favorites []*domain.Favorite
	for scanner.Next() {
		var favorite domain.Favorite
		err := scanner.Scan(&favorite.TweetID, &favorite.Username, &favorite.ID)
		if err != nil {
			sr.logging.Errorln(err)
			return nil, err
		}
		favorites = append(favorites, &favorite)
	}

	if err := scanner.Err(); err != nil {
		sr.logging.Errorln(err)
		return nil, err
	}

	return favorites, nil
}

func (sr *TweetRepo) GetTweetImage(ctx context.Context, id string) ([]byte, error) {
	ctx, span := sr.tracer.Start(ctx, "TweetStore.GetTweetImage")
	defer span.End()

	sr.logging.Infoln("Store: tweetImage reached")

	scanner := sr.session.Query(`SELECT * FROM tweet_image WHERE tweet_id = ?`, id).Iter().Scanner()

	var byteImage []byte
	for scanner.Next() {
		err := scanner.Scan(nil, &byteImage)
		if err != nil {
			sr.logging.Errorln(err)
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		sr.logging.Errorln(err)
		return nil, err
	}
	return byteImage, nil
}

func (sr *TweetRepo) Retweet(ctx context.Context, tweetID string, username string) (*gocql.UUID, int, error) {
	ctx, span := sr.tracer.Start(ctx, "TweetStore.Retweet")
	defer span.End()

	sr.logging.Infoln("Store: retweet reached")

	id, err := gocql.ParseUUID(tweetID)
	if err != nil {
		return nil, 500, nil
	}

	scanner := sr.session.Query(`SELECT * FROM retweet WHERE tweet_id = ? AND username = ?`, id.String(), username).Iter().Scanner()

	var retweets []*domain.Retweet

	for scanner.Next() {
		var retweet domain.Retweet
		err = scanner.Scan(&retweet.TweetID, &retweet.Username, &retweet.ID)
		if err != nil {
			sr.logging.Errorln(err)
			return nil, 502, err
		}
		retweets = append(retweets, &retweet)
	}

	if len(retweets) != 0 {
		sr.logging.Errorln("no retweets")
		return nil, 406, fmt.Errorf(errors.RetweetAlreadyExist)
	}

	scanner = sr.session.Query(`SELECT * FROM tweet WHERE id = ?`, id.String()).Iter().Scanner()

	var tweetUsername string
	var tweets []*domain.Tweet
	for scanner.Next() {
		var tweet domain.Tweet
		err := scanner.Scan(&tweet.ID, &tweet.CreatedAt, &tweet.Advertisement, &tweet.FavoriteCount, &tweet.Favorited,
			&tweet.Image, &tweet.OwnerUsername, &tweet.RetweetCount, &tweet.Retweeted, &tweet.Text, &tweet.Username)
		tweetUsername = tweet.Username
		if err != nil {
			sr.logging.Errorln(err)
			return nil, 500, err
		}

		tweets = append(tweets, &tweet)
	}

	if len(tweets) == 0 {
		sr.logger.Println("No such tweet")
		return nil, 500, nil
	}

	retweetCount := tweets[0].RetweetCount + 1
	retweeted := true
	createdAt := tweets[0].CreatedAt

	insert := fmt.Sprintf("INSERT INTO %s "+"(id, tweet_id, username) "+"VALUES (?, ?, ?)", COLLECTION_RETWEET)
	retweetID, _ := gocql.RandomUUID()

	err = sr.session.Query(
		insert, retweetID.String(), tweets[0].ID.String(), username).Exec()
	if err != nil {
		sr.logging.Errorln(err)
		return nil, 502, err
	}

	err = sr.session.Query(
		`UPDATE tweet SET retweeted=?, retweet_count=? WHERE id=? AND created_at=?`, retweeted, retweetCount, id.String(), createdAt).Exec()

	if err != nil {
		sr.logger.Println(err)
		return nil, 502, err
	}

	err = sr.session.Query(
		`UPDATE tweets_by_user SET retweeted=?, retweet_count=? WHERE username=? AND created_at=?`,
		retweeted, retweetCount, tweetUsername, createdAt).Exec()

	if err != nil {
		sr.logging.Errorln(err)
		return nil, 502, err
	}

	thisTweet := tweets[0]

	newID, err := gocql.RandomUUID()
	if err != nil {
		sr.logging.Errorln(err)
		return nil, 502, err
	}

	timeNow := time.Now().Unix()

	insertTweet := fmt.Sprintf("INSERT INTO %s "+
		"(id, created_at, favorite_count, favorited, retweet_count, retweeted, text, username, owner_username, image, advertisement) "+
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", COLLECTION)

	insertByUser := fmt.Sprintf("INSERT INTO %s "+
		"(id, created_at, favorite_count, favorited, retweet_count, retweeted, text, username, owner_username, image, advertisement) "+
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", COLLECTION_BY_USER)

	err = sr.session.Query(insertTweet,
		newID, timeNow, 0, false, 0, false, thisTweet.Text, username, thisTweet.Username, thisTweet.Image, false).Exec()

	err = sr.session.Query(insertByUser,
		newID, timeNow, 0, false, 0, false, thisTweet.Text, username, thisTweet.Username, thisTweet.Image, false).Exec()

	if err != nil {
		sr.logging.Errorln(err)
		return nil, 502, err
	}

	return &newID, 200, nil

}

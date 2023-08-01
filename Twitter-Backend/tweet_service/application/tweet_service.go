package application

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gocql/gocql"
	"github.com/nats-io/nats.go"
	"github.com/sony/gobreaker"
	events "github.com/zjalicf/twitter-clone-common/common/saga/create_event"
	"go.opentelemetry.io/otel/trace"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"tweet_service/domain"
)

var (
	followServiceHost = os.Getenv("FOLLOW_SERVICE_HOST")
	followServicePort = os.Getenv("FOLLOW_SERVICE_PORT")
)

type TweetService struct {
	store          domain.TweetStore
	tracer         trace.Tracer
	cache          domain.TweetCache
	cb             *gobreaker.CircuitBreaker
	nastConnection *nats.Conn
	elasticClient  *elasticsearch.Client
}

func NewTweetService(store domain.TweetStore, cache domain.TweetCache, tracer trace.Tracer, natsConnection *nats.Conn, elasticClient *elasticsearch.Client) *TweetService {
	return &TweetService{
		store:          store,
		cache:          cache,
		cb:             CircuitBreaker(),
		tracer:         tracer,
		nastConnection: natsConnection,
		elasticClient:  elasticClient,
	}
}

func (service *TweetService) GetAll(ctx context.Context) ([]domain.Tweet, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.GetAll")
	defer span.End()

	return service.store.GetAll(ctx)
}

func (service *TweetService) GetOne(ctx context.Context, tweetID string) (*domain.Tweet, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.GetOne")
	defer span.End()

	return service.store.GetOne(ctx, tweetID)
}

func (service *TweetService) GetTweetsByUser(ctx context.Context, username string) ([]*domain.Tweet, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.GetTweetsByUser")
	defer span.End()

	return service.store.GetTweetsByUser(ctx, username)
}

func (service *TweetService) GetFeedByUser(ctx context.Context, token string) (*domain.FeedData, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.GetFeedByUser")
	defer span.End()

	followServiceEndpoint := fmt.Sprintf("http://%s:%s/feedInfo", followServiceHost, followServicePort)
	followServiceRequest, _ := http.NewRequest("GET", followServiceEndpoint, nil)
	followServiceRequest.Header.Add("Authorization", token)
	bodyBytes, err := service.cb.Execute(func() (interface{}, error) {

		responseFservice, err := http.DefaultClient.Do(followServiceRequest)
		if err != nil {
			return nil, fmt.Errorf("FollowServiceError")
		}

		defer responseFservice.Body.Close()

		responseBodyBytes, err := io.ReadAll(responseFservice.Body)
		if err != nil {
			log.Printf("error in readAll: %s", err.Error())
			return nil, err
		}

		var feedInfo domain.FeedInfo
		err = json.Unmarshal(responseBodyBytes, &feedInfo)
		if err != nil {
			log.Printf("error in unmarshal: %s", err.Error())
			return nil, err
		}

		return feedInfo, nil
	})

	if err != nil {
		return nil, err
	}
	feedInfo := bodyBytes.(domain.FeedInfo)
	feed, err := service.store.GetPostsFeedByUser(ctx, feedInfo.Usernames)
	if err != nil {
		log.Println("Line 99: ", err)
		return nil, err
	}

	if len(feedInfo.AdIds) == 0 {
		return &domain.FeedData{
			Feed: feed,
			Ads:  nil,
		}, nil
	}

	ads, err := service.store.GetRecommendAdsForUser(ctx, feedInfo.AdIds)
	if err != nil {
		log.Printf("Error in getting recommend ads for user: %s", err.Error())
		return nil, err
	}

	return &domain.FeedData{
		Feed: feed,
		Ads:  ads,
	}, nil
}

func (service *TweetService) saveImage(ctx context.Context, tweetID gocql.UUID, imageBytes []byte) error {
	ctx, span := service.tracer.Start(ctx, "TweetService.saveImage")
	defer span.End()

	return service.store.SaveImage(ctx, tweetID, imageBytes)
}

func (service *TweetService) GetLikesByTweet(ctx context.Context, tweetID string) ([]*domain.Favorite, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.GetLikesByTweet")
	defer span.End()

	return service.store.GetLikesByTweet(ctx, tweetID)
}

func (service *TweetService) Post(ctx context.Context, tweet *domain.Tweet, username string, image *[]byte) (*domain.Tweet, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.Post")
	defer span.End()

	tweet.ID, _ = gocql.RandomUUID()

	tweet.Image = false
	if len(*image) != 0 {
		log.Printf("USLO U SLIKU")
		err := service.saveImage(ctx, tweet.ID, *image)
		if err != nil {
			return nil, err
		}

		err = service.cache.PostCacheData(ctx, tweet.ID.String(), image)
		if err != nil {
			return nil, err
		}
		tweet.Image = true
	}
	tweet.CreatedAt = time.Now().Unix()
	tweet.Favorited = false
	tweet.FavoriteCount = 0
	tweet.Retweeted = false
	tweet.RetweetCount = 0
	tweet.Username = username

	return service.store.Post(ctx, tweet)
}

func (service *TweetService) Favorite(ctx context.Context, id string, username string, isAd bool) (int, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.Favorite")
	defer span.End()

	status, err := service.store.Favorite(ctx, id, username)
	if err != nil {
		return status, err
	}

	if isAd {
		event := events.Event{
			TweetID:   id,
			Type:      "",
			Timestamp: time.Now().Unix(),
			Timespent: 0,
		}
		if status == 200 {
			event.Type = "Unliked"
		} else {
			event.Type = "Liked"
		}
		if err != nil {
			return status, err
		}
	}

	return status, nil
}

func (service *TweetService) GetTweetImage(ctx context.Context, id string) (*[]byte, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.GetTweetImage")
	defer span.End()

	cachedImage, _ := service.cache.GetCachedValue(ctx, id)

	if cachedImage != nil {
		return cachedImage, nil
	}

	image, err := service.store.GetTweetImage(ctx, id)
	if err != nil {
		return nil, err
	}

	err = service.cache.PostCacheData(ctx, id, &image)
	if err != nil {
		log.Printf("POST REDIS ERR: %s", err.Error())
		return nil, err
	}
	return &image, nil
}

func (service *TweetService) Retweet(ctx context.Context, id string, username string) (int, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.Retweet")
	defer span.End()

	tweet, err := service.store.GetOne(ctx, id)
	if err != nil {
		return 500, err
	}

	newUUID, status, err := service.store.Retweet(ctx, id, username)
	if err != nil {
		return status, err
	}

	if tweet.Image {
		image, err := service.store.GetTweetImage(ctx, tweet.ID.String())
		if err != nil {
			return 500, err
		}

		err = service.saveImage(ctx, *newUUID, image)
		if err != nil {
			log.Printf("Error in saving image of root tweet in retweet in TweetService.Retweet: %s", err.Error())
			return 500, err
		}
	}

	return status, nil
}

func CircuitBreaker() *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name:        "cb",
			MaxRequests: 1,
			Timeout:     time.Millisecond,
			Interval:    0,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 3
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				log.Printf("Circuit Breaker '%s' changed from '%s' to '%s'\n", name, from, to)
			},
		},
	)
}

func (service *TweetService) SubscribeToNats(natsConnection *nats.Conn) {

	//subscribe to chanel with name DELETE_TWEET
	_, err := natsConnection.QueueSubscribe(os.Getenv("DELETE_TWEET"), "queue-tweet-group", func(msg *nats.Msg) {

		var tweetID string
		err := json.Unmarshal(msg.Data, &tweetID)
		if err != nil {
			log.Println("Error in unmarshal JSON!")
			return
		}

		tweet, err := service.GetOne(context.Background(), tweetID)
		if err != nil {
			log.Println("Tweet with that id not exist")
			return
		}

		deleted := false
		err = service.store.DeleteOneTweet(tweet)
		if err == nil {
			deleted = true
		}

		dataToSend, err := json.Marshal(&deleted)
		if err != nil {
			log.Println("Error in marshaling json")
			return
		}

		err = natsConnection.Publish(msg.Reply, dataToSend)
		if err != nil {
			log.Printf("Error in publish response: %s", err.Error())
			return
		}

	})

	if err != nil {
		log.Printf("Error in receiving message: %s", err.Error())
		return
	}

}

func (service *TweetService) DeleteTweet(tweetID string) error {

	log.Println("USLO U DELET TWEET SERVICE LAYER")
	log.Println("TWEETID: ", tweetID)

	tweet, err := service.GetOne(context.Background(), tweetID)
	if err != nil {
		log.Println("Error in 321: ", err)
		return err
	}
	log.Println(tweet)

	err = service.store.DeleteOneTweet(tweet)
	if err != nil {
		log.Println("Error in 328: ", err)
		return err
	}

	return nil
}

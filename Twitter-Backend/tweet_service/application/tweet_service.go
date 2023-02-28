package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"tweet_service/domain"
	"github.com/gocql/gocql"
	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	events "github.com/zjalicf/twitter-clone-common/common/saga/create_event"
	"go.opentelemetry.io/otel/trace"
)

var (
	followServiceHost = os.Getenv("FOLLOW_SERVICE_HOST")
	followServicePort = os.Getenv("FOLLOW_SERVICE_PORT")
)

type TweetService struct {
	store        domain.TweetStore
	tracer       trace.Tracer
	cache        domain.TweetCache
	cb           *gobreaker.CircuitBreaker
	orchestrator *CreateEventOrchestrator
	logging      *logrus.Logger
}

func NewTweetService(store domain.TweetStore, cache domain.TweetCache, tracer trace.Tracer, orchestrator *CreateEventOrchestrator, logging *logrus.Logger) *TweetService {
	return &TweetService{
		store:        store,
		cache:        cache,
		cb:           CircuitBreaker(),
		orchestrator: orchestrator,
		tracer:       tracer,
		logging: logging,
	}
}

func (service *TweetService) GetAll(ctx context.Context) ([]domain.Tweet, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.GetAll")
	defer span.End()

	service.logging.Infoln("TweetService.GetAll : getAll service reached")

	return service.store.GetAll(ctx)
}

func (service *TweetService) GetOne(ctx context.Context, tweetID string) (*domain.Tweet, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.GetOne")
	defer span.End()

	service.logging.Infoln("TweetService.GetOne : getOne service reached")

	return service.store.GetOne(ctx, tweetID)
}

func (service *TweetService) GetTweetsByUser(ctx context.Context, username string) ([]*domain.Tweet, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.GetTweetsByUser")
	defer span.End()
	service.logging.Infoln("TweetService.getTweetsbyUser : tweetsByUser service reached")

	return service.store.GetTweetsByUser(ctx, username)
}

func (service *TweetService) GetFeedByUser(ctx context.Context, token string) (*domain.FeedData, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.GetFeedByUser")
	defer span.End()

	service.logging.Infoln("Feed : feed service reached")
	followServiceEndpoint := fmt.Sprintf("http://%s:%s/feedInfo", followServiceHost, followServicePort)
	followServiceRequest, _ := http.NewRequest("GET", followServiceEndpoint, nil)
	followServiceRequest.Header.Add("Authorization", token)
	bodyBytes, err := service.cb.Execute(func() (interface{}, error) {

		responseFservice, err := http.DefaultClient.Do(followServiceRequest)
		if err != nil {
			service.logging.Errorln("follow service error")
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
		service.logging.Errorln("error getting feed by user")
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

	service.logging.Infoln("saveimage : saveimage service reached")

	return service.store.SaveImage(ctx, tweetID, imageBytes)
}

func (service *TweetService) GetLikesByTweet(ctx context.Context, tweetID string) ([]*domain.Favorite, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.GetLikesByTweet")
	defer span.End()

	service.logging.Infoln("TweetService : likesbytweet service reached")

	return service.store.GetLikesByTweet(ctx, tweetID)
}

func (service *TweetService) Post(ctx context.Context, tweet *domain.Tweet, username string, image *[]byte) (*domain.Tweet, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.Post")
	defer span.End()

	service.logging.Infoln("TweetService : likesbytweet service reached")

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

	service.logging.Infoln("TweetService : favorite service reached")

	status, err := service.store.Favorite(ctx, id, username)
	if err != nil {
		service.logging.Errorln("TweetService : favorite error")
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
		log.Println(event.Timestamp)
		err = service.orchestrator.Start(ctx, event)
		if err != nil {
			return status, err
		}
	}

	return status, nil
}

func (service *TweetService) TimeSpentOnAd(ctx context.Context, timespent *domain.Timespent) error {
	ctx, span := service.tracer.Start(ctx, "TweetService.TimeSpentOnAd")
	defer span.End()

	service.logging.Infoln("TimeSpentOnAd : timespent service reached")

	event := events.Event{
		TweetID:   timespent.TweetID,
		Type:      "Timespent",
		Timestamp: time.Now().Unix(),
		Timespent: timespent.Timespent,
	}

	err := service.orchestrator.Start(ctx, event)
	if err != nil {
		service.logging.Errorln("orchestrator failure")
		return err
	}

	return nil
}

func (service *TweetService) GetTweetImage(ctx context.Context, id string) (*[]byte, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.GetTweetImage")
	defer span.End()

	service.logging.Infoln("TweetService : getTweetImage service reached")

	cachedImage, _ := service.cache.GetCachedValue(ctx, id)

	if cachedImage != nil {
		return cachedImage, nil
	}

	image, err := service.store.GetTweetImage(ctx, id)
	if err != nil {
		service.logging.Errorln("TweetService : gettweetImage error")
		return nil, err
	}

	err = service.cache.PostCacheData(ctx, id, &image)
	if err != nil {
		log.Printf("POST REDIS ERR: %s", err.Error())
		return nil, err
	}
	return &image, nil
}

func (service *TweetService) ViewProfileFromAd(ctx context.Context, tweetID domain.TweetID) error {
	ctx, span := service.tracer.Start(ctx, "TweetService.ViewProfileFromAd")
	defer span.End()

	service.logging.Infoln("TweetService : viewProfileFromAd service reached")

	event := events.Event{
		TweetID:   tweetID.ID,
		Type:      "ViewCount",
		Timestamp: time.Now().Unix(),
		Timespent: 0,
	}

	err := service.orchestrator.Start(ctx, event)
	if err != nil {
		service.logging.Errorln("TweetService : orchestrator fail")
		return err
	}
	return nil

}

func (service *TweetService) Retweet(ctx context.Context, id string, username string) (int, error) {
	ctx, span := service.tracer.Start(ctx, "TweetService.Retweet")
	defer span.End()

	service.logging.Infoln("TweetService : retweet service reached")

	tweet, err := service.store.GetOne(ctx, id)
	if err != nil {
		service.logging.Errorln("TweetService : getOne failed")
		return 500, err
	}

	newUUID, status, err := service.store.Retweet(ctx, id, username)
	if err != nil {
		return status, err
	}

	if tweet.Image {
		image, err := service.store.GetTweetImage(ctx, tweet.ID.String())
		if err != nil {
			service.logging.Errorln("TweetService : getTweetImage error")
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

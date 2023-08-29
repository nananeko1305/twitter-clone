package application

import (
	"context"
	"encoding/json"
	"firebase.google.com/go/messaging"
	"fmt"
	"follow_service/domain"
	"follow_service/errors"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/zjalicf/twitter-clone-common/common/saga/create_user"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"time"
)

type FollowService struct {
	store          domain.FollowRequestStore
	tracer         trace.Tracer
	nastConnection *nats.Conn
	messaging      *messaging.Client
}

func NewFollowService(store domain.FollowRequestStore, tracer trace.Tracer, natsConnection *nats.Conn, messaging *messaging.Client) *FollowService {
	return &FollowService{
		store:          store,
		tracer:         tracer,
		nastConnection: natsConnection,
		messaging:      messaging,
	}
}

func (service *FollowService) FollowExist(ctx context.Context, followRequest *domain.FollowRequest) (bool, error) {
	ctx, span := service.tracer.Start(ctx, "FollowService.FollowExist")
	defer span.End()

	return service.store.FollowExist(ctx, followRequest)
}

func (service *FollowService) GetFeedInfoOfUser(ctx context.Context, username string) (*domain.FeedInfo, error) {
	ctx, span := service.tracer.Start(ctx, "FollowService.GetFeedInfoOfUser")
	defer span.End()

	followings, err := service.store.GetFollowingsOfUser(ctx, username)
	if err != nil {
		return nil, err
	}

	followings = append(followings, username)

	recommendAds, err := service.store.GetRecommendAdsId(ctx, username)
	if err != nil {
		return nil, err
	}

	return &domain.FeedInfo{
		Usernames: followings,
		AdIds:     recommendAds,
	}, nil
}

func (service *FollowService) GetFollowingsOfUser(ctx context.Context, username string) ([]string, error) {
	ctx, span := service.tracer.Start(ctx, "FollowService.GetFollowingsOfUser")
	defer span.End()

	followings, err := service.store.GetFollowingsOfUser(ctx, username)
	if err != nil {
		return nil, err
	}

	return followings, nil
}

func (service *FollowService) GetFollowersOfUser(ctx context.Context, username string) ([]string, error) {
	ctx, span := service.tracer.Start(ctx, "FollowService.GetFollowersOfUser")
	defer span.End()

	followings, err := service.store.GetFollowersOfUser(ctx, username)
	if err != nil {
		return nil, err
	}

	return followings, nil
}

func (service *FollowService) GetRequestsForUser(ctx context.Context, username string) ([]*domain.FollowRequest, error) {
	ctx, span := service.tracer.Start(ctx, "FollowService.GetRequestsForUser")
	defer span.End()

	return service.store.GetRequestsForUser(ctx, username)
}

func (service *FollowService) CreateRequest(ctx context.Context, request *domain.FollowRequest, username string, visibility bool) error {
	ctx, span := service.tracer.Start(ctx, "FollowService.CreateRequest")
	defer span.End()

	request.ID = uuid.New().String()
	request.Requester = username

	isExist, err := service.FollowExist(ctx, request)
	if err != nil {
		return err
	}

	if isExist {
		return fmt.Errorf("You already follow this user!")
	}

	if visibility {
		existing, err := service.store.GetRequestByRequesterReceiver(ctx, &request.Requester, &request.Receiver)
		if err != nil {
			if err.Error() == errors.ErrorRequestNotExists {
				request.Status = 1
				err = service.store.SaveRequest(ctx, request)
				if err != nil {
					return fmt.Errorf("Request not inserted in db")
				}
				return nil
			} else {
				return err
			}
		}

		existing.Status = 1
		err = service.store.UpdateRequest(ctx, existing)
		if err != nil {
			return fmt.Errorf("Request not inserted in db")
		}

	} else {
		err := service.store.SaveFollow(ctx, request)
		if err != nil {
			return err
		}
	}

	dataToSend, err := json.Marshal(username)
	if err != nil {
		log.Println("Error in marshaling json to send with NATS: ", err)
		return err
	}

	response, err := service.nastConnection.Request(os.Getenv("GET_FCM_TOKEN"), dataToSend, 5*time.Second)
	if err != nil {
		log.Println("Error with send request", err)
		return err
	}

	var fcmToken string
	err = json.Unmarshal(response.Data, &fcmToken)
	if err != nil {
		log.Println("Error in unmarshal json")
		return err
	}

	service.sendNotificationToDevice(fcmToken, username)

	return nil
}

func (service *FollowService) CreateUser(ctx context.Context, user *domain.User) error {
	ctx, span := service.tracer.Start(ctx, "FollowService.CreateUser")
	defer span.End()

	err := service.store.SaveUser(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (service *FollowService) AcceptRequest(ctx context.Context, id *string) error {
	ctx, span := service.tracer.Start(ctx, "FollowService.AcceptRequest")
	defer span.End()

	request, err := service.store.AcceptRequest(ctx, id)
	if err != nil {
		return fmt.Errorf(errors.ErrorInAcceptRequest)
	}

	err = service.store.SaveFollow(ctx, request)
	if err != nil {
		return fmt.Errorf(errors.ErrorInSaveFollow)
	}

	return nil

}

func (service *FollowService) DeleteUser(ctx context.Context, id *string) error {
	ctx, span := service.tracer.Start(ctx, "FollowService.AcceptRequest")
	defer span.End()

	return service.store.DeleteUser(ctx, id)
}

func (service *FollowService) DeclineRequest(ctx context.Context, id *string) error {
	ctx, span := service.tracer.Start(ctx, "FollowService.DeclineRequest")
	defer span.End()

	return service.store.DeclineRequest(ctx, id)
}

func (service *FollowService) HandleRequest(ctx context.Context, followRequest *domain.FollowRequest) error {
	ctx, span := service.tracer.Start(ctx, "FollowService.HandleRequest")
	defer span.End()

	return service.store.SaveRequest(ctx, followRequest)
}

func (service *FollowService) SaveAd(ctx context.Context, ad *domain.Ad) error {
	ctx, span := service.tracer.Start(ctx, "FollowService.SaveAd")
	defer span.End()

	return service.store.SaveAd(ctx, ad)
}

func (service *FollowService) GetRecommendationsByUsername(ctx context.Context, username string) ([]string, error) {
	ctx, span := service.tracer.Start(ctx, "FollowService.GetRecommendationsByUsername")
	defer span.End()

	countFollowings, err := service.store.CountFollowings(ctx, username)
	if err != nil {
		return nil, err
	}

	if countFollowings == 0 {
		recommendations, err := service.store.RecommendationWithoutFollowings(ctx, username, []string{})
		if err != nil {
			return nil, err
		}

		return recommendations, nil
	} else {
		var allRecommendations []string
		recommendations, err := service.store.RecommendWithFollowings(ctx, username)
		if err != nil {
			return nil, err
		}

		recommendations2, err := service.store.RecommendationWithoutFollowings(ctx, username, recommendations)
		if err != nil {
			return nil, err
		}

		allRecommendations = append(recommendations, recommendations2...)

		return allRecommendations, nil
	}

}

func (service *FollowService) UserToDomain(userIn create_user.User) domain.User {
	var user domain.User
	user.ID = userIn.ID.Hex()
	user.Age = userIn.Age
	user.Residence = userIn.Residence
	user.Username = userIn.Username
	user.Gender = string(userIn.Gender)
	if user.Age == 0 {
		user.Gender = ""
	}

	return user
}

func (service *FollowService) sendNotificationToDevice(token string, username string) error {

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: "TITLE",
			Body:  "" + username + " send you follow request!",
		},
	}

	_, err := service.messaging.Send(context.Background(), message)
	if err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}
	log.Println("Message sent")

	return nil
}

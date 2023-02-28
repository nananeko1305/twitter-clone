package application

import (
	"context"
	"fmt"
	"follow_service/domain"
	"follow_service/errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/zjalicf/twitter-clone-common/common/saga/create_user"
	"go.opentelemetry.io/otel/trace"
)

type FollowService struct {
	store   domain.FollowRequestStore
	tracer  trace.Tracer
	logging *logrus.Logger
}

func NewFollowService(store domain.FollowRequestStore, tracer trace.Tracer, logging *logrus.Logger) *FollowService {
	return &FollowService{
		store:   store,
		tracer:  tracer,
		logging: logging,
	}
}

func (service *FollowService) FollowExist(ctx context.Context, followRequest *domain.FollowRequest) (bool, error) {
	ctx, span := service.tracer.Start(ctx, "FollowService.FollowExist")
	defer span.End()

	service.logging.Infoln("FollowService.FollowExist : follow_exist reached")

	return service.store.FollowExist(ctx, followRequest)
}

func (service *FollowService) GetFeedInfoOfUser(ctx context.Context, username string) (*domain.FeedInfo, error) {
	ctx, span := service.tracer.Start(ctx, "FollowService.GetFeedInfoOfUser")
	defer span.End()

	service.logging.Infoln("FollowService.GetFeedInfoOfUser : GetFeedInfoOfUser reached")

	followings, err := service.store.GetFollowingsOfUser(ctx, username)
	if err != nil {
		service.logging.Errorf("FollowService.GetFollowingsOfUser.GetFollowingsOfUser() : %s", err)
		return nil, err
	}

	followings = append(followings, username)

	recommendAds, err := service.store.GetRecommendAdsId(ctx, username)
	if err != nil {
		service.logging.Errorf("FollowService.GetFollowingsOfUser.GetRecommendAdsId() : %s", err)
		return nil, err
	}

	service.logging.Infoln("FollowService.GetFeedInfoOfUser : GetFeedInfoOfUser successful")

	return &domain.FeedInfo{
		Usernames: followings,
		AdIds:     recommendAds,
	}, nil
}

func (service *FollowService) GetFollowingsOfUser(ctx context.Context, username string) ([]string, error) {
	ctx, span := service.tracer.Start(ctx, "FollowService.GetFollowingsOfUser")
	defer span.End()

	service.logging.Infoln("FollowService.GetFollowingsOfUser : GetFollowingsOfUser reached")

	followings, err := service.store.GetFollowingsOfUser(ctx, username)
	if err != nil {
		service.logging.Errorf("FollowService.GetFollowingsOfUser.GetFollowingsOfUser() : %s", err)
		return nil, err
	}

	service.logging.Infoln("FollowService.GetFollowingsOfUser : GetFollowingsOfUser successful")

	return followings, nil
}

func (service *FollowService) GetFollowersOfUser(ctx context.Context, username string) ([]string, error) {
	ctx, span := service.tracer.Start(ctx, "FollowService.GetFollowersOfUser")
	defer span.End()

	service.logging.Infoln("FollowService.GetFollowersOfUser : GetFollowersOfUser reached")

	followings, err := service.store.GetFollowersOfUser(ctx, username)
	if err != nil {
		service.logging.Errorf("FollowService.GetFollowersOfUser.GetFollowersOfUser() : %s", err)
		return nil, err
	}

	service.logging.Infoln("FollowService.GetFollowersOfUser : GetFollowersOfUser successful")

	return followings, nil
}

func (service *FollowService) GetRequestsForUser(ctx context.Context, username string) ([]*domain.FollowRequest, error) {
	ctx, span := service.tracer.Start(ctx, "FollowService.GetRequestsForUser")
	defer span.End()

	service.logging.Infoln("FollowService.FollowExist : follow_exist reached")

	return service.store.GetRequestsForUser(ctx, username)
}

func (service *FollowService) CreateRequest(ctx context.Context, request *domain.FollowRequest, username string, visibility bool) error {
	ctx, span := service.tracer.Start(ctx, "FollowService.CreateRequest")
	defer span.End()

	service.logging.Infoln("FollowService.CreateRequest : create_request reached")

	request.ID = uuid.New().String()
	request.Requester = username

	isExist, err := service.FollowExist(ctx, request)
	if err != nil {
		service.logging.Errorf("FollowService.CreateRequest.FollowExist() : %s", err)
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
					service.logging.Errorf("FollowService.CreateRequest.SaveRequest() : %s", err)
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
			service.logging.Errorf("FollowService.CreateRequest.UpdateRequest() : %s", err)
			return fmt.Errorf("Request not inserted in db")
		}

	} else {
		err := service.store.SaveFollow(ctx, request)
		if err != nil {
			service.logging.Errorf("FollowService.CreateRequest.SaveFollow() : %s", err)
			return err
		}
	}

	service.logging.Infoln("FollowService.CreateRequest : create_request successful")

	return nil
}

func (service *FollowService) CreateUser(ctx context.Context, user *domain.User) error {
	ctx, span := service.tracer.Start(ctx, "FollowService.CreateUser")
	defer span.End()

	service.logging.Infoln("FollowService.CreateUser : create_user reached")

	err := service.store.SaveUser(ctx, user)
	if err != nil {
		service.logging.Errorf("FollowService.CreateUser.SaveUser() : %s", err)
		return err
	}

	service.logging.Infoln("FollowService.CreateUser : create_user successful")

	return nil
}

func (service *FollowService) AcceptRequest(ctx context.Context, id *string) error {
	ctx, span := service.tracer.Start(ctx, "FollowService.AcceptRequest")
	defer span.End()

	service.logging.Infoln("FollowService.AcceptRequest : accept_request reached")

	request, err := service.store.AcceptRequest(ctx, id)
	if err != nil {
		service.logging.Errorf("FollowService.AcceptRequest.AcceptRequest() : %s", err)
		return fmt.Errorf(errors.ErrorInAcceptRequest)
	}

	err = service.store.SaveFollow(ctx, request)
	if err != nil {
		service.logging.Errorf("FollowService.AcceptRequest.SaveFollow() : %s", err)
		return fmt.Errorf(errors.ErrorInSaveFollow)
	}

	service.logging.Infoln("FollowService.AcceptRequest : accept_request successful")

	return nil

}

func (service *FollowService) DeleteUser(ctx context.Context, id *string) error {
	ctx, span := service.tracer.Start(ctx, "FollowService.AcceptRequest")
	defer span.End()

	service.logging.Infoln("FollowService.DeleteUser : AcceptRequest reached")

	return service.store.DeleteUser(ctx, id)
}

func (service *FollowService) DeclineRequest(ctx context.Context, id *string) error {
	ctx, span := service.tracer.Start(ctx, "FollowService.DeclineRequest")
	defer span.End()

	service.logging.Infoln("FollowService.DeclineRequest : DeclineRequest reached")

	return service.store.DeclineRequest(ctx, id)
}

func (service *FollowService) HandleRequest(ctx context.Context, followRequest *domain.FollowRequest) error {
	ctx, span := service.tracer.Start(ctx, "FollowService.HandleRequest")
	defer span.End()

	service.logging.Infoln("FollowService.HandleRequest : HandleRequest reached")

	return service.store.SaveRequest(ctx, followRequest)
}

func (service *FollowService) SaveAd(ctx context.Context, ad *domain.Ad) error {
	ctx, span := service.tracer.Start(ctx, "FollowService.SaveAd")
	defer span.End()

	service.logging.Infoln("FollowService.SaveAd : SaveAd reached")

	return service.store.SaveAd(ctx, ad)
}

func (service *FollowService) GetRecommendationsByUsername(ctx context.Context, username string) ([]string, error) {
	ctx, span := service.tracer.Start(ctx, "FollowService.GetRecommendationsByUsername")
	defer span.End()

	service.logging.Infoln("FollowService.GetRecommendationsByUsername : GetRecommendationsByUsername reached")

	countFollowings, err := service.store.CountFollowings(ctx, username)
	if err != nil {
		service.logging.Errorf("FollowService.GetRecommendationsByUsername.CountFollowings() : %s", err)
		return nil, err
	}

	if countFollowings == 0 {
		recommendations, err := service.store.RecommendationWithoutFollowings(ctx, username, []string{})
		if err != nil {
			service.logging.Errorf("FollowService.GetRecommendationsByUsername.RecommendationWithoutFollowings() : %s", err)
			return nil, err
		}

		service.logging.Infoln("FollowService.GetRecommendationsByUsername : GetRecommendationsByUsername successful")

		return recommendations, nil
	} else {
		var allRecommendations []string
		recommendations, err := service.store.RecommendWithFollowings(ctx, username)
		if err != nil {
			service.logging.Errorf("FollowService.GetRecommendationsByUsername.RecommendWithFollowings() : %s", err)
			return nil, err
		}

		recommendations2, err := service.store.RecommendationWithoutFollowings(ctx, username, recommendations)
		if err != nil {
			service.logging.Errorf("FollowService.GetRecommendationsByUsername.RecommendationWithoutFollowings() : %s", err)
			return nil, err
		}

		allRecommendations = append(recommendations, recommendations2...)

		service.logging.Infoln("FollowService.GetRecommendationsByUsername : GetRecommendationsByUsername successful")

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

package domain

import "context"

type FollowRequestStore interface {
	SaveUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id *string) error
	GetRequestsForUser(ctx context.Context, username string) ([]*FollowRequest, error)
	GetRequestByRequesterReceiver(ctx context.Context, requester, receiver *string) (*FollowRequest, error)
	GetFollowingsOfUser(ctx context.Context, username string) ([]string, error)
	GetFollowersOfUser(ctx context.Context, username string) ([]string, error)
	SaveRequest(ctx context.Context, followRequest *FollowRequest) error
	SaveFollow(ctx context.Context, request *FollowRequest) error
	AcceptRequest(ctx context.Context, id *string) (*FollowRequest, error)
	DeclineRequest(ctx context.Context, id *string) error
	FollowExist(ctx context.Context, followRequest *FollowRequest) (bool, error)
	UpdateRequest(ctx context.Context, request *FollowRequest) error
	SaveAd(ctx context.Context, ad *Ad) error
	GetRecommendAdsId(ctx context.Context, username string) ([]string, error)
	CountFollowings(ctx context.Context, username string) (int, error)
	RecommendWithFollowings(ctx context.Context, username string) ([]string, error)
	RecommendationWithoutFollowings(ctx context.Context, username string, recommends []string) ([]string, error)
}

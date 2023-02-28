package application

import (
	"context"
	"fmt"
	"log"
	"user_service/domain"
	"user_service/errors"

	"github.com/sirupsen/logrus"
	"github.com/zjalicf/twitter-clone-common/common/saga/create_user"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

type UserService struct {
	store  domain.UserStore
	tracer trace.Tracer
	logging *logrus.Logger
}

func NewUserService(store domain.UserStore, tracer trace.Tracer, logging *logrus.Logger) *UserService {
	return &UserService{
		store:  store,
		tracer: tracer,
		logging: logging,
	}
}

func (service *UserService) Get(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	ctx, span := service.tracer.Start(ctx, "UserService.Get")
	defer span.End()

	service.logging.Infoln("UserService.Get : get service reached")

	return service.store.Get(ctx, id)
}

func (service *UserService) DoesEmailExist(ctx context.Context, email string) (string, error) {
	ctx, span := service.tracer.Start(ctx, "UserService.DoesEmailExist")
	defer span.End()
	service.logging.Infoln("UserService.DoesEmailExist : email service reached")

	user, err := service.store.GetByEmail(ctx, email)
	if err != nil {
		service.logging.Errorln("UserService failed getByEmail")
		return "", err
	}

	return user.ID.Hex(), nil
}

func (service *UserService) GetAll(ctx context.Context) ([]*domain.User, error) {
	ctx, span := service.tracer.Start(ctx, "UserService.GetAll")
	defer span.End()
	service.logging.Infoln("UserService.GetAll : getAll service reached")

	return service.store.GetAll(ctx)
}

func (service *UserService) GetOneUser(ctx context.Context, username string) (*domain.User, error) {
	ctx, span := service.tracer.Start(ctx, "UserService.GetOneUser")
	defer span.End()

	service.logging.Infoln("UserService.GetOneUser : getOneUser service reached")

	retUser, err := service.store.GetOneUser(ctx, username)
	if err != nil {
		service.logging.Errorln("UserService.GetOneUser : failed to query")
		return nil, fmt.Errorf("user not found")
	}
	return retUser, nil
}

func (service *UserService) Register(ctx context.Context, user *domain.User) (*domain.User, error) {

	ctx, span := service.tracer.Start(ctx, "UserService.Register")
	defer span.End()
	service.logging.Infoln("UserService.Register : register service reached")

	validatedUser, err := validateUserType(user)
	if err != nil {
		log.Println(errors.ValidationError)
		return nil, fmt.Errorf(errors.ValidationError)
	}
	if validatedUser.UserType == "Business" {
		validatedUser.Privacy = false
	} else {
		validatedUser.Privacy = true
	}

	retUser, err := service.store.Post(ctx, validatedUser)
	if err != nil {
		service.logging.Errorln("UserService failed to insert user into db")
		log.Println(errors.DatabaseError)
		return nil, fmt.Errorf(errors.DatabaseError)
	}

	return retUser, nil

}

func (service *UserService) ChangeUserVisibility(ctx context.Context, userID string) error {
	ctx, span := service.tracer.Start(ctx, "UserService.ChangeUserVisibility")
	defer span.End()

	service.logging.Infoln("UserService.ChangeVisibility : visibility service reached")

	primitiveID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		service.logging.Errorln("UserService failed to parse userID")
		log.Println("Primitive ID parsing error.")
		return err
	}

	user, err := service.store.Get(ctx, primitiveID)
	if err != nil {
		service.logging.Errorln("UserService failed to get user")
		log.Printf("Getting user by id error: %s", err.Error())
		return fmt.Errorf(errors.UserNotFound)
	}

	user.Privacy = !user.Privacy
	err = service.store.UpdateUser(ctx, user)
	if err != nil {
		service.logging.Errorln("UserService failed to update user")
		log.Printf("Updating user error in service: %s", err.Error())
		return err
	}

	return nil
}

func (service *UserService) DeleteUserByID(ctx context.Context, id primitive.ObjectID) error {
	ctx, span := service.tracer.Start(ctx, "UserService.DeleteUserByID")
	defer span.End()

	service.logging.Infoln("UserService.Delete : delete service reached")

	return service.store.DeleteUserByID(ctx, id)
}

func validateUserType(user *domain.User) (*domain.User, error) {

	business := isBusiness(user)
	regular := isRegular(user)

	if business && regular {
		return nil, fmt.Errorf("invalid user format")
	} else if business {
		user.UserType = domain.Business
		return user, nil
	} else if regular {
		user.UserType = domain.Regular
		return user, nil
	}

	return nil, fmt.Errorf("invalid user data")
}

func isBusiness(user *domain.User) bool {
	if len(user.CompanyName) >= 3 &&
		len(user.Website) >= 3 &&
		len(user.Email) >= 3 &&
		len(user.Username) >= 3 {
		return true
	}

	return false
}

func isRegular(user *domain.User) bool {
	if len(user.Firstname) >= 3 &&
		len(user.Lastname) >= 3 &&
		len(user.Gender) >= 3 &&
		user.Age >= 1 &&
		len(user.Residence) >= 3 &&
		len(user.Username) >= 3 {
		return true
	}

	return false
}

func (service *UserService) UserToDomain(userIn create_user.User) domain.User {
	var user domain.User
	user.ID = userIn.ID
	user.Firstname = userIn.Firstname
	user.Lastname = userIn.Lastname
	if userIn.Gender == "Male" {
		user.Gender = "Male"
	} else {
		user.Gender = "Female"
	}
	user.Age = userIn.Age
	user.Residence = userIn.Residence
	user.Email = userIn.Email
	user.Username = userIn.Username
	if userIn.UserType == "Regular" {
		user.UserType = "Regular"
	} else {
		user.UserType = "Business"
	}
	user.Privacy = userIn.Visibility
	user.CompanyName = userIn.CompanyName
	user.Website = userIn.Website

	return user
}

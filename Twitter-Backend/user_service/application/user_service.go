package application

import (
	"context"
	"fmt"
	"github.com/zjalicf/twitter-clone-common/common/saga/create_user"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
	"log"
	"user_service/domain"
	"user_service/errors"
)

type UserService struct {
	store          domain.UserStore
	tracer         trace.Tracer
	olivereElastic domain.UserElasticStore
}

func NewUserService(store domain.UserStore, olivereElastic domain.UserElasticStore, tracer trace.Tracer) *UserService {
	return &UserService{
		store:          store,
		tracer:         tracer,
		olivereElastic: olivereElastic,
	}
}

func (service *UserService) Get(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	ctx, span := service.tracer.Start(ctx, "UserService.Get")
	defer span.End()

	return service.store.Get(ctx, id)
}

func (service *UserService) DoesEmailExist(ctx context.Context, email string) (string, error) {
	ctx, span := service.tracer.Start(ctx, "UserService.DoesEmailExist")
	defer span.End()

	user, err := service.store.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	return user.ID.Hex(), nil
}

func (service *UserService) GetAll(ctx context.Context) ([]*domain.User, error) {
	ctx, span := service.tracer.Start(ctx, "UserService.GetAll")
	defer span.End()

	return service.store.GetAll(ctx)
}

func (service *UserService) GetOneUser(ctx context.Context, username string) (*domain.User, error) {
	ctx, span := service.tracer.Start(ctx, "UserService.GetOneUser")
	defer span.End()

	retUser, err := service.store.GetOneUser(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return retUser, nil
}

func (service *UserService) Register(ctx context.Context, user *domain.User) (*domain.User, error) {

	ctx, span := service.tracer.Start(ctx, "UserService.Register")
	defer span.End()

	validatedUser, err := validateUserType(user)
	if err != nil {
		log.Println(errors.ValidationError)
		return nil, fmt.Errorf(errors.ValidationError)
	}
	if validatedUser.UserType == "Business" {
		validatedUser.Privacy = false
	} else if validatedUser.UserType == "Regular" {
		validatedUser.Privacy = true
	} else {
		validatedUser.Privacy = false
	}

	retUser, err := service.store.Post(ctx, validatedUser)
	if err != nil {
		log.Println(errors.DatabaseError)
		return nil, fmt.Errorf(errors.DatabaseError)
	}

	if retUser.UserType != "Admin" {
		err = service.olivereElastic.Post(*retUser)
		if err != nil {
			return nil, err
		}
	}

	return retUser, nil

}

func (service *UserService) ChangeUserVisibility(ctx context.Context, userID string) error {
	ctx, span := service.tracer.Start(ctx, "UserService.ChangeUserVisibility")
	defer span.End()

	primitiveID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println("Primitive ID parsing error.")
		return err
	}

	user, err := service.store.Get(ctx, primitiveID)
	if err != nil {
		log.Printf("Getting user by id error: %s", err.Error())
		return fmt.Errorf(errors.UserNotFound)
	}

	user.Privacy = !user.Privacy
	err = service.store.UpdateUser(ctx, user)
	if err != nil {
		log.Printf("Updating user error in service: %s", err.Error())
		return err
	}

	return nil
}

func (service *UserService) DeleteUserByID(ctx context.Context, id primitive.ObjectID) error {
	ctx, span := service.tracer.Start(ctx, "UserService.DeleteUserByID")
	defer span.End()

	return service.store.DeleteUserByID(ctx, id)
}

func validateUserType(user *domain.User) (*domain.User, error) {

	if user.UserType == domain.Admin {
		user.UserType = domain.Admin
		return user, nil
	}

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
	} else if userIn.UserType == "Business" {
		user.UserType = "Business"
	} else {
		user.UserType = "Admin"
	}
	user.Privacy = userIn.Visibility
	user.CompanyName = userIn.CompanyName
	user.Website = userIn.Website

	return user
}

// Elasticsearch

//func (service *UserService) Test(username string) ([]*domain.User, error) {
//	return service.olivereElastic.
//}

func (service *UserService) CheckIndex() {
	service.olivereElastic.CheckIndex()
}

func (service *UserService) Search(search domain.Search) ([]*domain.User, error) {
	return service.olivereElastic.Search(search)
}

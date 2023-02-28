package application

import (
	"auth_service/domain"
	"context"
	events "github.com/zjalicf/twitter-clone-common/common/saga/create_user"
	saga "github.com/zjalicf/twitter-clone-common/common/saga/messaging"
	"go.opentelemetry.io/otel/trace"
	"log"
)

type CreateUserOrchestrator struct {
	commandPublisher saga.Publisher
	replySubscriber  saga.Subscriber
	tracer           trace.Tracer
}

func NewCreateUserOrchestrator(publisher saga.Publisher, subscriber saga.Subscriber, tracer trace.Tracer) (*CreateUserOrchestrator, error) {
	orchestrator := &CreateUserOrchestrator{
		commandPublisher: publisher,
		replySubscriber:  subscriber,
		tracer:           tracer,
	}
	err := orchestrator.replySubscriber.Subscribe(orchestrator.handle)
	if err != nil {
		return nil, err
	}
	return orchestrator, nil
}

func (o *CreateUserOrchestrator) Start(ctx context.Context, user *domain.User) error {
	ctx, span := o.tracer.Start(ctx, "AuthService.CreateUserOrchestrator.Start")
	defer span.End()

	var gender events.Gender
	var userType events.UserType

	if user.Gender == "Male" {
		gender = "Male"
	} else {
		gender = "Female"
	}

	if user.UserType == "Regular" {
		userType = "Regular"
	} else {
		userType = "Business"
	}

	user1 := events.User{
		ID:          user.ID,
		Firstname:   user.Firstname,
		Lastname:    user.Lastname,
		Gender:      gender,
		Age:         user.Age,
		Residence:   user.Residence,
		Email:       user.Email,
		Username:    user.Username,
		Password:    user.Password,
		UserType:    userType,
		Visibility:  user.Visibility,
		CompanyName: user.CompanyName,
		Website:     user.Website,
	}

	event := &events.CreateUserCommand{
		User: user1,
		Type: events.UpdateAuth,
	}

	return o.commandPublisher.Publish(event)
}

func (o *CreateUserOrchestrator) handle(reply *events.CreateUserReply) {
	command := events.CreateUserCommand{User: reply.User}
	command.Type = o.nextCommandType(*reply)
	if command.Type != events.UnknownCommand {
		_ = o.commandPublisher.Publish(command)
	}
}

func (o *CreateUserOrchestrator) nextCommandType(reply events.CreateUserReply) events.CreateUserCommandType {
	switch reply.Type {
	case events.AuthUpdated:
		log.Println("AUTH UPDATED")
		return events.UpdateUsers
	case events.UsersUpdated:
		log.Println("USERS UPDATED")
		return events.UpdateGraph
	case events.GraphUpdated:
		log.Println("GRAPH UPDATED")
		return events.SendMail
	case events.MailSent:
		log.Println("MAIL SENT")
		return events.UnknownCommand
	case events.MailFailed:
		log.Println("MAIL FAILED")
		return events.RollbackFollow
	case events.FollowFailed:
		log.Println("FOLLOW FAILED")
		return events.RollbackUsers
	case events.UsersFailed:
		log.Println("USERS FAILED")
		return events.RollbackAuth
	case events.UnknownReply:
		return events.UnknownCommand
	default:
		return events.UnknownCommand
	}
}

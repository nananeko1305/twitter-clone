package handlers

import (
	"context"
	"fmt"
	events "github.com/zjalicf/twitter-clone-common/common/saga/create_user"
	saga "github.com/zjalicf/twitter-clone-common/common/saga/messaging"
	"go.opentelemetry.io/otel/trace"
	"user_service/application"
)

type CreateUserCommandHandler struct {
	userService       *application.UserService
	replyPublisher    saga.Publisher
	commandSubscriber saga.Subscriber
	tracer            trace.Tracer
}

func NewCreateUserCommandHandler(userService *application.UserService, publisher saga.Publisher, subscriber saga.Subscriber, tracer trace.Tracer) (*CreateUserCommandHandler, error) {
	o := &CreateUserCommandHandler{
		userService:       userService,
		replyPublisher:    publisher,
		commandSubscriber: subscriber,
		tracer:            tracer,
	}
	err := o.commandSubscriber.Subscribe(o.handle)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (handler *CreateUserCommandHandler) handle(command *events.CreateUserCommand) {

	user := handler.userService.UserToDomain(command.User)
	reply := events.CreateUserReply{User: command.User}

	switch command.Type {

	case events.UpdateUsers:

		_, err := handler.userService.Register(context.Background(), &user)
		if err != nil {
			reply.Type = events.UsersFailed
		} else {
			reply.Type = events.UsersUpdated

		}

	case events.RollbackUsers:
		_ = handler.userService.DeleteUserByID(context.Background(), user.ID)
		reply.Type = events.UsersFailed
		fmt.Println("Rollback users")

	default:
		reply.Type = events.UnknownReply
	}

	if reply.Type != events.UnknownReply {
		_ = handler.replyPublisher.Publish(reply)
	}
}

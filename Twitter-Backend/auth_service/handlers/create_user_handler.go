package handlers

import (
	"auth_service/application"
	"context"
	"fmt"
	events "github.com/zjalicf/twitter-clone-common/common/saga/create_user"
	saga "github.com/zjalicf/twitter-clone-common/common/saga/messaging"
	"log"
)

type CreateUserCommandHandler struct {
	authService       *application.AuthService
	publisher         saga.Publisher
	commandSubscriber saga.Subscriber
}

func NewCreateUserCommandHandler(authService *application.AuthService, publisher saga.Publisher, subscriber saga.Subscriber) (*CreateUserCommandHandler, error) {
	o := &CreateUserCommandHandler{
		authService:       authService,
		publisher:         publisher,
		commandSubscriber: subscriber,
	}
	//prijava za slusanje komandi
	err := o.commandSubscriber.Subscribe(o.handle)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (handler *CreateUserCommandHandler) handle(command *events.CreateUserCommand) {

	user := handler.authService.UserToDomain(command.User)
	reply := events.CreateUserReply{User: command.User}

	switch command.Type {

	case events.UpdateAuth:
		reply.Type = events.AuthUpdated

	case events.SendMail:
		err := handler.authService.SendMail(context.Background(), &user)
		if err != nil {
			log.Printf("Failed to send mail: %s", err.Error())
			reply.Type = events.MailFailed
		} else {
			reply.Type = events.MailSent

		}

	case events.RollbackAuth:
		_ = handler.authService.DeleteUserByID(context.Background(), user.ID)
		reply.Type = events.UnknownReply
		fmt.Println("Rollback auth")

	default:
		reply.Type = events.UnknownReply
	}

	if reply.Type != events.UnknownReply {
		_ = handler.publisher.Publish(reply)
	}
}

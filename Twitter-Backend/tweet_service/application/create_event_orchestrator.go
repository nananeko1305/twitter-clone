package application

import (
	"context"
	events "github.com/zjalicf/twitter-clone-common/common/saga/create_event"
	saga "github.com/zjalicf/twitter-clone-common/common/saga/messaging"
	"go.opentelemetry.io/otel/trace"
	"log"
)

type CreateEventOrchestrator struct {
	commandPublisher saga.Publisher
	replySubscriber  saga.Subscriber
	tracer           trace.Tracer
}

func NewCreateEventOrchestrator(publisher saga.Publisher, subscriber saga.Subscriber, tracer trace.Tracer) (*CreateEventOrchestrator, error) {
	o := &CreateEventOrchestrator{
		commandPublisher: publisher,
		replySubscriber:  subscriber,
		tracer:           tracer,
	}
	err := o.replySubscriber.Subscribe(o.handle)
	if err != nil {
		log.Printf("Error in create orchestrator NewCreateEventOrchestrator(): %s", err.Error())
		return nil, err
	}
	return o, nil
}

func (o *CreateEventOrchestrator) Start(ctx context.Context, event events.Event) error {
	ctx, span := o.tracer.Start(ctx, "Orchestrator Starting")
	defer span.End()

	log.Printf("Starting orchestrator with eventType: %s", event.Type)

	eventPush := events.CreateEventCommand{
		Event: event,
		Type:  events.UpdateCassandra,
	}

	_, span1 := o.tracer.Start(ctx, "Event published")
	defer span1.End()

	return o.commandPublisher.Publish(eventPush)
}

func (o *CreateEventOrchestrator) handle(reply *events.CreateEventReply) {
	log.Printf("Orkestrator primio reply: %s", reply.Type)
	command := events.CreateEventCommand{Event: reply.Event}
	command.Type = o.nextCommandType(*reply)
	if command.Type != events.UnknownCommand {
		log.Printf("Orkestrator salje command: %s", command.Type)
		_ = o.commandPublisher.Publish(command)
	}
}

func (o *CreateEventOrchestrator) nextCommandType(reply events.CreateEventReply) events.CreateEventCommandType {

	log.Println(reply.Type)

	switch reply.Type {
	case events.CassandraUpadated:
		log.Println("CassandraUpdated")
		return events.UpdateMongo

	case events.MongoUpdated:
		log.Println("MongoUpdated")
		return events.UnknownCommand

	case events.UnknownReply:
		return events.UnknownCommand

	default:
		log.Println("unknown command nextCommandType orchestrator")
		return events.UnknownCommand
	}
}

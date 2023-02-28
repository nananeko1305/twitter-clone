package handlers

import (
	events "github.com/zjalicf/twitter-clone-common/common/saga/create_event"
	saga "github.com/zjalicf/twitter-clone-common/common/saga/messaging"
	"log"
	"report_service/application"
)

type CreateReportCommandHandler struct {
	reportService     *application.ReportService
	replyPublisher    saga.Publisher
	commandSubscriber saga.Subscriber
}

func NewCreateEventCommandHandler(reportService *application.ReportService, replyPublisher saga.Publisher, subscriber saga.Subscriber) (*CreateReportCommandHandler, error) {
	o := &CreateReportCommandHandler{
		reportService:     reportService,
		replyPublisher:    replyPublisher,
		commandSubscriber: subscriber,
	}
	//prijava za slusanje komandi
	err := o.commandSubscriber.Subscribe(o.handle)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// hendlovanje komandama
func (handler *CreateReportCommandHandler) handle(command *events.CreateEventCommand) {
	reply := events.CreateEventReply{Event: command.Event}

	switch command.Type {

	case events.UpdateCassandra:
		log.Println("Primljen event update cassandra")
		handler.reportService.CreateEvent(command.Event)
		reply.Type = events.CassandraUpadated

	case events.UpdateMongo:
		log.Println("Primljen event update mongo")
		handler.reportService.CreateReport(&command.Event)
		reply.Type = events.MongoUpdated

	default:
		log.Println("Unknown reply report handler")
		reply.Type = events.UnknownReply
	}

	if reply.Type != events.UnknownReply {
		log.Println("event publish in report handler")
		_ = handler.replyPublisher.Publish(reply)
		log.Printf("event is %s", reply.Type)

	}
}

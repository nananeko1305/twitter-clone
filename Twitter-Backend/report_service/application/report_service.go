package application

import (
	"context"
	"github.com/sirupsen/logrus"
	events "github.com/zjalicf/twitter-clone-common/common/saga/create_event"
	"go.opentelemetry.io/otel/trace"
	"log"
	"report_service/domain"
	"time"
)

type ReportService struct {
	eventStore  domain.EventStore
	reportStore domain.ReportStore
	tracer      trace.Tracer
	logger      *logrus.Logger
}

func NewReportService(eventStore domain.EventStore, reportStore domain.ReportStore, tracer trace.Tracer, logger *logrus.Logger) *ReportService {
	return &ReportService{
		eventStore:  eventStore,
		reportStore: reportStore,
		tracer:      tracer,
		logger:      logger,
	}
}

func (service *ReportService) CreateEvent(event events.Event) {
	eventOut := EventToDomain(event)
	_, err := service.eventStore.CreateEvent(context.TODO(), &eventOut)
	if err != nil {
		log.Printf("Error in report_service CreateEvent()", err.Error())
		return
	}
}

func (service *ReportService) CreateReport(event *events.Event) {
	ctx, span := service.tracer.Start(context.TODO(), "ReportService.CreateReport")
	defer span.End()

	service.logger.Infoln("ReportService.CreateReport : Create Report service reached")

	if event.Type == "Timespent" {
		domainEvent := EventToDomain(*event)
		timeSpendDaily, err := service.eventStore.GetTimespentDailyEvents(ctx, &domainEvent)
		if err != nil {
			service.logger.Errorf("Error in getting daily spent: %s", err.Error())
			return
		}
		timeSpendMonthly, err := service.eventStore.GetTimespentMonthlyEvents(ctx, &domainEvent)
		if err != nil {
			service.logger.Errorf("Error in getting monthly spent: %s", err.Error())
			return
		}

		event.DailySpent = timeSpendDaily
		event.MonthlySpent = timeSpendMonthly
	}

	log.Printf("Timedaily :%s , TimeMonthly: %s", event.DailySpent, event.MonthlySpent)
	thisT := time.Unix(event.Timestamp, 0)
	localTime := time.Date(thisT.Year(), thisT.Month(), thisT.Day(), thisT.Hour()+1, 0, 0, 0, time.Local)

	dailyUnix := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, time.Local).Unix() //dailyRepDate
	monthlyUnix := time.Date(localTime.Year(), localTime.Month(), 1, 0, 0, 0, 0, time.Local).Unix()             //monthlyRepDate
	_, err := service.reportStore.CreateReport(ctx, event, monthlyUnix, dailyUnix)
	if err != nil {
		service.logger.Errorf("Error in report_service CreateReport(): %s", err.Error())
		return
	} else {
		log.Println("Succesfull updated report")
	}

}

func (service *ReportService) GetReportForAd(ctx context.Context, tweetID string, reportType string, date int64) (*domain.Report, error) {
	ctx, span := service.tracer.Start(context.TODO(), "ReportService.GetReportForAd")
	defer span.End()

	service.logger.Infoln("ReportService.GetReportForAd : GetReportForAd service reached")

	result, err := service.reportStore.GetReportForAd(ctx, tweetID, reportType, date)
	if err != nil {
		service.logger.Errorf("Error in ReportService GetReportForAd: %s", err.Error())
		return nil, err
	}
	return result, nil
}

func EventToDomain(event events.Event) domain.Event {

	return domain.Event{
		TweetID:      event.TweetID,
		Type:         event.Type,
		Timestamp:    event.Timestamp,
		Timespent:    event.Timespent,
		DailySpent:   event.DailySpent,
		MonthlySpent: event.MonthlySpent,
	}
}

package store

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	events "github.com/zjalicf/twitter-clone-common/common/saga/create_event"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/trace"
	"log"
	"report_service/domain"
)

const (
	DATABASE           = "report_mongo"
	COLLECTION_DAILY   = "daily_reports"
	COLLECTION_MONTHLY = "monthly_reports"
	UnknownEventError  = "Unknown event type"
)

type ReportMongoDBStore struct {
	dailyReports   *mongo.Collection
	monthlyReports *mongo.Collection
	tracer         trace.Tracer
	logging        *logrus.Logger
}

func NewReportMongoDBStore(client *mongo.Client, tracer trace.Tracer, logging *logrus.Logger) domain.ReportStore {
	dailyReports := client.Database(DATABASE).Collection(COLLECTION_DAILY)
	monthlyReports := client.Database(DATABASE).Collection(COLLECTION_MONTHLY)
	return &ReportMongoDBStore{
		dailyReports:   dailyReports,
		monthlyReports: monthlyReports,
		tracer:         tracer,
		logging:        logging,
	}
}

func (store *ReportMongoDBStore) GetReportForAd(ctx context.Context, tweetID string, reportType string, timestamp int64) (*domain.Report, error) {
	ctx, span := store.tracer.Start(ctx, "ReportMongoDBStore.GetReportForAd")
	defer span.End()

	store.logging.Infoln("ReportStore.GetReportForAd : reached Get Report For Ad in store")

	if reportType == "daily" {

		result, err := store.filterOneDaily(bson.M{"tweet_id": tweetID, "timestamp": timestamp})
		if err != nil {
			store.logging.Errorf("ReportStore.GetReportForAd.FilterOneDaily() : %s", err)
			return nil, err
		}
		return result, nil

	} else if reportType == "monthly" {

		result, err := store.filterOneMonthly(bson.M{"tweet_id": tweetID, "timestamp": timestamp})
		if err != nil {
			store.logging.Errorf("ReportStore.GetReportForAd.FilterOneMonthly() : %s", err)
			return nil, err
		}
		return result, nil

	}
	return nil, nil
}

func (store *ReportMongoDBStore) CreateReport(ctx context.Context, event *events.Event,
	monthlyUnix, dailyUnix int64) (*events.Event, error) {
	ctx, span := store.tracer.Start(ctx, "ReportMongoDBStore.CreateReport")
	defer span.End()

	store.logging.Infoln("ReportStore.CreateReport : reached CreateReport in store")

	oneDaily, err := store.filterOneDaily(bson.M{"tweet_id": event.TweetID, "timestamp": dailyUnix})
	if err != nil {
		store.logging.Errorf("Error in ReportMongoStore.filterOneDaily(), filterOneDaily: %s", err.Error())
		report := domain.Report{
			ID:          primitive.NewObjectID(),
			TweetID:     event.TweetID,
			Timestamp:   dailyUnix,
			LikeCount:   0,
			UnlikeCount: 0,
			ViewCount:   0,
			Timespent:   0,
		}

		if event.Type == "Liked" {
			report.LikeCount++
		} else if event.Type == "Unliked" {
			report.UnlikeCount++
		} else if event.Type == "Timespent" {
			report.Timespent = int(event.DailySpent)
		} else if event.Type == "ViewCount" {
			report.ViewCount++
		} else {
			return nil, fmt.Errorf(UnknownEventError)
		}

		_, err = store.dailyReports.InsertOne(ctx, report)
		if err != nil {
			store.logging.Errorf("Error in ReportMongoStore, daily_report: %s", err.Error())
			return nil, err
		}
	} else {

		//like update
		if event.Type == "Liked" {
			oneDaily.LikeCount = oneDaily.LikeCount + 1
		} else if event.Type == "Unliked" {
			//unline update
			oneDaily.UnlikeCount = oneDaily.UnlikeCount + 1
		} else if event.Type == "Timespent" {
			oneDaily.Timespent = int(event.DailySpent)
		} else if event.Type == "ViewCount" {
			//view update
			oneDaily.ViewCount = oneDaily.ViewCount + 1
		} else {
			return nil, fmt.Errorf(UnknownEventError)
		}

		_, err = store.dailyReports.UpdateOne(context.TODO(), bson.M{"_id": oneDaily.ID}, bson.M{"$set": oneDaily})
		if err != nil {
			store.logging.Errorf("Error in report_mongodb CreateReport() Unlike monthly: %s", err.Error())
			return nil, err
		}

	}
	//monthly
	oneMonthly, err := store.filterOneMonthly(bson.M{"tweet_id": event.TweetID, "timestamp": monthlyUnix})
	if err != nil {
		store.logging.Errorf("Error in ReportMongoStore.filterOneMonthly(), filterOneMonthly: %s", err.Error())
		report := domain.Report{
			ID:          primitive.NewObjectID(),
			TweetID:     event.TweetID,
			Timestamp:   monthlyUnix,
			LikeCount:   0,
			UnlikeCount: 0,
			ViewCount:   0,
			Timespent:   0,
		}

		if event.Type == "Liked" {
			report.LikeCount++
		} else if event.Type == "Unliked" {
			report.UnlikeCount++
		} else if event.Type == "Timespent" {
			report.Timespent = int(event.MonthlySpent)
		} else if event.Type == "ViewCount" {
			report.ViewCount++
		} else {
			return nil, fmt.Errorf(UnknownEventError)
		}
		_, err = store.monthlyReports.InsertOne(ctx, report)
		if err != nil {
			store.logging.Errorf("Error in ReportMongoStore, monthly_reporst: %s", err.Error())
			return nil, err
		}
	} else {

		//like update
		if event.Type == "Liked" {
			oneMonthly.LikeCount = oneMonthly.LikeCount + 1

		} else if event.Type == "Unliked" {
			//unline update
			oneMonthly.UnlikeCount = oneMonthly.UnlikeCount + 1

		} else if event.Type == "Timespent" {
			oneMonthly.Timespent = int(event.MonthlySpent)
		} else if event.Type == "ViewCount" {
			//view update
			oneMonthly.ViewCount = oneMonthly.ViewCount + 1
		} else {
			return nil, fmt.Errorf(UnknownEventError)
		}
		_, err = store.monthlyReports.UpdateOne(context.TODO(), bson.M{"_id": oneMonthly.ID}, bson.M{"$set": oneMonthly})
		if err != nil {
			store.logging.Errorf("Error in report_mongodb CreateReport() Like monthly: %s", err.Error())
			return nil, err
		}

	}

	return event, nil
}

func (store *ReportMongoDBStore) filterDaily(filter interface{}) ([]*domain.Report, error) {
	cursor, err := store.dailyReports.Find(context.TODO(), filter)
	defer cursor.Close(context.TODO())

	if err != nil {
		store.logging.Errorf("Error in report_mongodb filter() Unlike daily: %s", err.Error())
		return nil, err
	}
	return decode(cursor)
}

func (store *ReportMongoDBStore) filterOneDaily(filter interface{}) (user *domain.Report, err error) {
	result := store.dailyReports.FindOne(context.TODO(), filter)
	err = result.Decode(&user)
	return
}

func (store *ReportMongoDBStore) filterOneMonthly(filter interface{}) (user *domain.Report, err error) {
	result := store.monthlyReports.FindOne(context.TODO(), filter)
	err = result.Decode(&user)
	return
}

func decode(cursor *mongo.Cursor) (reports []*domain.Report, err error) {
	for cursor.Next(context.TODO()) {
		var report domain.Report
		err = cursor.Decode(&report)
		if err != nil {
			log.Printf("Error in report_mongodb decode(): %s", err.Error())
			return
		}
		reports = append(reports, &report)
	}
	err = cursor.Err()
	return
}

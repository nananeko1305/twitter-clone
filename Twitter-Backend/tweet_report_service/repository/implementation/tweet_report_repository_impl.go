package implementation

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"tweet_report_service/domain"
	"tweet_report_service/repository"
)

const (
	DATABASE      = "tweet_reports"
	REPORTS       = "reports"
	REPORTS_COUNT = "report_count"
)

type TweetReportRepositoryImpl struct {
	reports     *mongo.Collection
	reportCount *mongo.Collection
}

func NewTweetReportRepositoryImpl(client *mongo.Client) repository.TweetReportRepository {
	reports := client.Database(DATABASE).Collection(REPORTS)
	reportCount := client.Database(DATABASE).Collection(REPORTS_COUNT)

	return TweetReportRepositoryImpl{
		reports:     reports,
		reportCount: reportCount,
	}

}

func (repository TweetReportRepositoryImpl) Get() ([]*domain.TweetReport, error) {

	find, err := repository.reports.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}

	var reports []*domain.TweetReport

	for find.Next(context.Background()) {
		var report *domain.TweetReport
		err := find.Decode(&report)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	return reports, nil
}

func (repository TweetReportRepositoryImpl) Post(report domain.TweetReport) error {

	_, err := repository.reports.InsertOne(context.Background(), report)
	if err != nil {
		return err
	}
	return nil
}

func (repository TweetReportRepositoryImpl) Put(report domain.TweetReport) error {

	filter := bson.M{"_id": report.Id}
	update := bson.M{"$set": bson.M{"status": report.Status}}

	_, err := repository.reports.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (repository TweetReportRepositoryImpl) Delete(id primitive.ObjectID) error {

	_, err := repository.reports.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return err
	}
	return nil
}

func (repository TweetReportRepositoryImpl) IsReportedByUser(report domain.TweetReport) error {

	filter := bson.M{"username": report.Username, "tweetID": report.TweetID}
	log.Println(filter)

	result := repository.reports.FindOne(context.Background(), filter)

	err := result.Decode(&report)
	if err != nil {
		return err
	}

	return nil
}

func (repository TweetReportRepositoryImpl) GetReportCountForTweet(tweetID string) (*domain.ReportCount, error) {

	filter := bson.M{"tweet_id": tweetID}
	result := repository.reportCount.FindOne(context.Background(), filter)
	var reportCount domain.ReportCount
	err := result.Decode(&reportCount)
	if err != nil {
		log.Println("Error in finding one ReportCount: ", err)
		return nil, err
	}

	return &reportCount, nil
}

func (repository TweetReportRepositoryImpl) CreateNewReportCountForTweet(reportCount domain.ReportCount) error {

	_, err := repository.reportCount.InsertOne(context.Background(), reportCount)
	if err != nil {
		return err
	}

	return nil
}

func (repository TweetReportRepositoryImpl) UpdateReportCountForTweet(reportCount *domain.ReportCount) error {

	filter := bson.M{"tweet_id": reportCount.TweetID}
	update := bson.M{"$set": bson.M{"count": reportCount.Count}}
	_, err := repository.reportCount.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

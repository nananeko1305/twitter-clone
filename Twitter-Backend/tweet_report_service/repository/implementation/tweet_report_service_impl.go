package implementation

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"tweet_report_service/domain"
	"tweet_report_service/repository"
)

const (
	DATABASE   = "tweet_reports"
	COLLECTION = "reports"
)

type TweetReportRepositoryImpl struct {
	reports *mongo.Collection
}

func NewTweetReportRepositoryImpl(client *mongo.Client) repository.TweetReportRepository {
	reports := client.Database(DATABASE).Collection(COLLECTION)

	return TweetReportRepositoryImpl{
		reports: reports,
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

func (repository TweetReportRepositoryImpl) Delete(id primitive.ObjectID) error {

	_, err := repository.reports.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return err
	}
	return nil
}

func (repository TweetReportRepositoryImpl) IsReportedByUser(report domain.TweetReport) error {

	filter := bson.M{"username": report.Username, "tweetID": report.TweetID}

	result := repository.reports.FindOne(context.Background(), filter)

	err := result.Decode(&report)
	if err != nil {
		return err
	}

	return nil
}

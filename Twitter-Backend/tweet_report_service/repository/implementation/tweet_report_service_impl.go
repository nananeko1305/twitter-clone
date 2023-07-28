package implementation

import (
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

func (t TweetReportRepositoryImpl) Create(report domain.TweetReport) error {
	//TODO implement me
	panic("implement me")
}

func (t TweetReportRepositoryImpl) Read(report domain.TweetReport) error {
	//TODO implement me
	panic("implement me")
}

func (t TweetReportRepositoryImpl) Delete(id primitive.ObjectID) error {
	//TODO implement me
	panic("implement me")
}

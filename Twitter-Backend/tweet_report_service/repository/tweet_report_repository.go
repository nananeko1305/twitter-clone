package repository

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"tweet_report_service/domain"
)

type TweetReportRepository interface {
	Post(report domain.TweetReport) error
	Get() ([]*domain.TweetReport, error)
	Put(report domain.TweetReport) error
	Delete(id primitive.ObjectID) error
	IsReportedByUser(report domain.TweetReport) error
	UpdateReportCountForTweet(reportCount *domain.ReportCount) error
	CreateNewReportCountForTweet(reportCount domain.ReportCount) error
	GetReportCountForTweet(tweetID string) (*domain.ReportCount, error)
}

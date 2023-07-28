package repository

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"tweet_report_service/domain"
)

type TweetReportRepository interface {
	Create(report domain.TweetReport) error
	Read(report domain.TweetReport) error
	Delete(id primitive.ObjectID) error
}

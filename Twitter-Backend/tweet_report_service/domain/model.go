package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TweetReport struct {
	Id       primitive.ObjectID `json:"id" bson:"_id"`
	Username string             `json:"username" bson:"username"`
	TweetID  string             `json:"tweetID" bson:"tweetID"`
	Reason   Reason             `json:"reason" bson:"reason"`
	Status   Status             `json:"status" bson:"status"`
}

type Reason string

const (
	Spam          Reason = "Spam"
	Hate          Reason = "Hate"
	Inappropriate Reason = "InappropriateContent"
	VerbalAbuse   Reason = "VerbalAbuse"
	DontLikeIt    Reason = "DontLikeIt"
)

type Status string

const (
	Pending  Status = "PENDING"
	Accepted Status = "ACCEPTED"
	Declined Status = "DECLINED"
)

type ReportCount struct {
	ID      primitive.ObjectID `bson:"_id"`
	TweetID string             `bson:"tweet_id"`
	Count   int                `bson:"count"`
}

func CheckReason(reason Reason) bool {
	switch reason {
	case Spam, Hate, Inappropriate, DontLikeIt, VerbalAbuse:
		return true
	default:
		return false
	}
}

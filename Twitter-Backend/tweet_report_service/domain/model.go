package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TweetReport struct {
	Id       primitive.ObjectID `json:"id" bson:"_id"`
	Username string             `json:"username" bson:"username"`
	TweetID  string             `json:"tweetID" bson:"tweetID"`
	Reason   Reason             `json:"reason" bson:"reason"`
}

type Reason string

const (
	Spam          Reason = "Spam"
	Hate          Reason = "Hate"
	Inappropriate Reason = "InappropriateContent"
	VerbalAbuse   Reason = "VerbalAbuse"
	DontLikeIt    Reason = "DontLikeIt"
)

func CheckReason(reason Reason) bool {
	switch reason {
	case Spam, Hate, Inappropriate, DontLikeIt, VerbalAbuse:
		return true
	default:
		return false
	}
}

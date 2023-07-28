package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type TweetReport struct {
	id       primitive.ObjectID
	username string
	tweetID  string
	reason   Reason
}

type Reason string

const (
	InappropriateContent = "InappropriateContent"
	VerbalAbuse          = "VerbalAbuse"
	Other                = "Other"
)

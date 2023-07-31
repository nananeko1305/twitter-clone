package service

import (
	"encoding/json"
	"errors"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"os"
	"time"
	"tweet_report_service/domain"
	"tweet_report_service/repository"
)

type TweetReportService struct {
	repository     repository.TweetReportRepository
	natsConnection *nats.Conn
}

func NewTweetReportService(repository repository.TweetReportRepository, natsConnection *nats.Conn) *TweetReportService {
	return &TweetReportService{
		repository:     repository,
		natsConnection: natsConnection,
	}
}

func (service *TweetReportService) Get() ([]*domain.TweetReport, error) {

	reports, err := service.repository.Get()
	if err != nil {
		log.Println("Error in database: ", err)
		return nil, err
	}

	return reports, nil
}

func (service *TweetReportService) Post(report domain.TweetReport) error {

	if !domain.CheckReason(report.Reason) {
		return errors.New("invalid reason")
	}

	err := service.repository.IsReportedByUser(report)
	if err != nil {
		report.Id = primitive.NewObjectID()
		report.Status = domain.Pending

		err := service.repository.Post(report)
		if err != nil {
			log.Println("Error in collection reports: ", err)
			return err
		}

		return nil
	} else {
		return errors.New("user already reported this tweet")
	}

}

func (service *TweetReportService) Put(report domain.TweetReport) (*domain.TweetReport, error) {

	switch report.Status {

	case domain.Accepted:
		report.Status = domain.Accepted
		err := service.UpdateReportCount(report)
		if err != nil {
			return nil, err
		}

	case domain.Declined:
		report.Status = domain.Declined

	default:
		report.Status = domain.Pending
	}

	err := service.repository.Put(report)
	if err != nil {
		log.Println("Error in database: ", err)
		return nil, err
	}
	return &report, nil

}

func (service *TweetReportService) Delete(id string) error {

	primitiveID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Error in converting from string to primitiveObjectID: ", err)
		return err
	}

	err = service.repository.Delete(primitiveID)
	if err != nil {
		log.Println("Error in database: ", err)
		return err
	}

	return nil
}

func (service *TweetReportService) UpdateReportCount(report domain.TweetReport) error {

	result, err := service.repository.GetReportCountForTweet(report.TweetID)

	if err != nil {
		var newReportCount domain.ReportCount
		newReportCount.ID = report.Id
		newReportCount.TweetID = report.TweetID
		newReportCount.Count = 1

		err := service.repository.CreateNewReportCountForTweet(newReportCount)
		if err != nil {
			log.Println("Error in collection reportCount: ", err)
			return err
		}
		return nil
	}

	//return to 2!!!!!!
	if result.Count == 0 {

		//pozivanje message brokera

		dataToSend, err := json.Marshal(result.TweetID)
		if err != nil {
			log.Println("Error in marshaling json to send with NATS: ", err)
			return err
		}

		response, err := service.natsConnection.Request(os.Getenv("DELETE_TWEET"), dataToSend, 5*time.Second)
		if err != nil {
			return err
		}

		var tweetDeleted bool
		err = json.Unmarshal(response.Data, &tweetDeleted)
		if err != nil {
			log.Println("Error in unmarshal json")
			return err
		}

		if tweetDeleted {
			log.Println("TWEET DELETED")
		} else {
			log.Println("TWEET IS NOT DELETED")
			return errors.New("tweet is not deleted")
		}

		result.Count = result.Count + 1
	} else if result.Count < 2 {
		result.Count = result.Count + 1
	} else {
		return nil
	}
	err = service.repository.UpdateReportCountForTweet(result)
	if err != nil {
		log.Println("Error in updating report_count: ", err)
		return err
	}
	return nil
}

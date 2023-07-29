package service

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"tweet_report_service/domain"
	"tweet_report_service/repository"
)

type TweetReportService struct {
	repository repository.TweetReportRepository
}

func NewTweetReportService(repository repository.TweetReportRepository) *TweetReportService {
	return &TweetReportService{
		repository: repository,
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

		err := service.repository.Post(report)
		if err != nil {
			log.Println("Error in database: ", err)
			return err
		}
		return nil
	} else {
		return errors.New("user already reported this tweet")
	}

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

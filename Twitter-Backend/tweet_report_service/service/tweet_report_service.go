package service

import "tweet_report_service/repository"

type TweetReportService struct {
	repository *repository.TweetReportRepository
}

func NewTweetReportService(repository *repository.TweetReportRepository) *TweetReportService {
	return &TweetReportService{
		repository: repository,
	}
}

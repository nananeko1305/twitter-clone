package controller

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"tweet_report_service/service"
)

type TweetReportController struct {
	tweetReportService *service.TweetReportService
}

func NewTweetReportController(tweetReportService *service.TweetReportService) *TweetReportController {
	return &TweetReportController{
		tweetReportService: tweetReportService,
	}
}

func (controller *TweetReportController) InitRoutes(router *mux.Router) {

	router.HandleFunc("/get", controller.Get).Methods("GET")

	http.Handle("/", router)
	log.Println("Server started successfully!")

}

func (controller *TweetReportController) Get(response http.ResponseWriter, request *http.Request) {
	fmt.Println("Hello world")

	response.WriteHeader(http.StatusOK)
}

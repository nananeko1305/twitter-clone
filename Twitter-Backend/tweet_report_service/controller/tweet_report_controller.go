package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"tweet_report_service/domain"
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

	router.HandleFunc("/reports", controller.Get).Methods("GET")
	router.HandleFunc("/reports", controller.Post).Methods("POST")
	router.HandleFunc("/reports/{id}", controller.Delete).Methods("DELETE")

	http.Handle("/", router)
	log.Println("Server started successfully!")

}

func (controller *TweetReportController) Get(response http.ResponseWriter, request *http.Request) {

	reports, err := controller.tweetReportService.Get()
	if err != nil {
		response.WriteHeader(http.StatusOK)
		response.Write([]byte(err.Error()))
		return
	}

	jsonData, err := json.Marshal(reports)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(jsonData)
}

func (controller *TweetReportController) Post(response http.ResponseWriter, request *http.Request) {

	var report domain.TweetReport

	err := json.NewDecoder(request.Body).Decode(&report)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(err.Error()))
	}

	err = controller.tweetReportService.Post(report)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(err.Error()))
	}
}

func (controller *TweetReportController) Delete(response http.ResponseWriter, request *http.Request) {

	id := mux.Vars(request)["id"]
	err := controller.tweetReportService.Delete(id)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(err.Error()))
		return
	}

	response.WriteHeader(http.StatusOK)
}

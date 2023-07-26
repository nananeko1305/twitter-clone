package controller

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type TweetReportController struct {
}

func NewTweetReportController() *TweetReportController {
	return &TweetReportController{}
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

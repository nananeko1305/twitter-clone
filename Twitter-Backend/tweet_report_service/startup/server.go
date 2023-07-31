package startup

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"tweet_report_service/configs"
	"tweet_report_service/controller"
	"tweet_report_service/repository/implementation"
	"tweet_report_service/service"
	"tweet_report_service/startup/config"
)

type Server struct {
	config *config.Config
}

func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
	}
}

func (server *Server) Start() {

	natsConnection := configs.ConnectToNats(server.config)
	defer natsConnection.Close()

	mongoClient := configs.ConnectToMongoDB(server.config.DBHost, server.config.DBPort)

	tweetReportRepository := implementation.NewTweetReportRepositoryImpl(mongoClient)
	tweetReportService := service.NewTweetReportService(tweetReportRepository, natsConnection)
	tweetReportController := controller.NewTweetReportController(tweetReportService)

	router := mux.NewRouter()
	tweetReportController.InitRoutes(router)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", server.config.ServicePort),
		Handler: router,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Println(err)
		return
	}
}

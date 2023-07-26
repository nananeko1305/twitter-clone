package startup

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"tweet_report_service/controller"
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

	httpHandler := controller.NewTweetReportController()
	router := mux.NewRouter()
	httpHandler.InitRoutes(router)

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

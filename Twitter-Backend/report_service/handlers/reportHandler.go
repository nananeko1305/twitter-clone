package handlers

import (
	"github.com/casbin/casbin"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"report_service/application"
	"report_service/authorization"
	"report_service/domain"
	"strconv"
)

type KeyUser struct{}

type ReportHandler struct {
	service *application.ReportService
	store   *domain.EventStore
	tracer  trace.Tracer
	logging *logrus.Logger
}

func NewReportHandler(service *application.ReportService, tracer trace.Tracer, logging *logrus.Logger) *ReportHandler {
	return &ReportHandler{
		service: service,
		tracer:  tracer,
		logging: logging,
	}
}

func (handler *ReportHandler) Init(router *mux.Router) {
	reportEnforcer, err := casbin.NewEnforcerSafe("./auth_model.conf", "./policy.csv")
	handler.logging.Infoln("report_service : successful init of enforcer")
	if err != nil {
		log.Fatal(err)
	}

	router.HandleFunc("/{id}/{reportType}/{date}", handler.GetReportForAd).Methods("GET")
	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":8005", authorization.Authorizer(reportEnforcer)(router)))

}

func (handler *ReportHandler) GetReportForAd(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "AuthHandler.Register")
	defer span.End()

	handler.logging.Infoln("ReportHandler.GetReportForAd : Get Report For Ad endpoint reached")

	vars := mux.Vars(req)
	timestamp, _ := strconv.Atoi(vars["date"])

	log.Printf("TweetID : %s, reportType : %s, timestamp: %s", vars["id"], vars["reportType"], vars["date"])

	ad, err := handler.service.GetReportForAd(ctx, vars["id"], vars["reportType"], int64(timestamp))
	if err != nil {
		handler.logging.Errorf("ReportHandler.GetReportForAd : %s", err)
		http.Error(writer, "Error in handler GetReportForAd", http.StatusInternalServerError)
		return
	}
	jsonResponse(ad, writer)
	writer.WriteHeader(http.StatusOK)
}

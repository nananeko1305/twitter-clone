package handlers

import (
	"github.com/casbin/casbin"
	"github.com/gorilla/mux"
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
}

func NewReportHandler(service *application.ReportService, tracer trace.Tracer) *ReportHandler {
	return &ReportHandler{
		service: service,
		tracer:  tracer,
	}
}

func (handler *ReportHandler) Init(router *mux.Router) {
	reportEnforcer, err := casbin.NewEnforcerSafe("./auth_model.conf", "./policy.csv")
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

	vars := mux.Vars(req)
	timestamp, _ := strconv.Atoi(vars["date"])

	log.Printf("TweetID : %s, reportType : %s, timestamp: %s", vars["id"], vars["reportType"], vars["date"])

	ad, err := handler.service.GetReportForAd(ctx, vars["id"], vars["reportType"], int64(timestamp))
	if err != nil {
		http.Error(writer, "Error in handler GetReportForAd", http.StatusInternalServerError)
		return
	}
	jsonResponse(ad, writer)
	writer.WriteHeader(http.StatusOK)
}

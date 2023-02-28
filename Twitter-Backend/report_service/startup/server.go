package startup

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	saga "github.com/zjalicf/twitter-clone-common/common/saga/messaging"
	"github.com/zjalicf/twitter-clone-common/common/saga/messaging/nats"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"os"
	"os/signal"
	"report_service/application"
	"report_service/domain"
	"report_service/handlers"
	"report_service/startup/config"
	"report_service/store"
	"syscall"
	"time"
)

var Logger = logrus.New()

type Server struct {
	config *config.Config
}

const (
	QueueGroup  = "report_service"
	LogFilePath = "/app/logs/application.log"
)

func initLogger() {
	file, err := os.OpenFile(LogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logrus.Fatal(err)
	}

	Logger.SetOutput(file)

	rotationInterval := 24 * time.Hour
	ticker := time.NewTicker(rotationInterval)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			rotateLogs(file)
		}
	}()
}

func rotateLogs(file *os.File) {
	currentTime := time.Now().Format("2006-01-02_15-04-05")
	err := os.Rename("/app/logs/application.log", "/app/logs/application_"+currentTime+".log")
	if err != nil {
		Logger.Error(err)
	}
	file.Close()

	file, err = os.OpenFile("/app/logs/application.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		Logger.Error(err)
	}

	Logger.SetOutput(file)
}

func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
	}
}

func (server *Server) Start() {

	initLogger()

	mongoClient := server.initMongoClient()
	defer func(mongoClient *mongo.Client, ctx context.Context) {
		err := mongoClient.Disconnect(ctx)
		if err != nil {
			log.Printf("Error in server Start(): %s", err.Error())
		}
	}(mongoClient, context.Background())

	cfg := config.NewConfig()

	ctx := context.Background()
	exp, err := newExporter(cfg.JaegerAddress)
	if err != nil {
		log.Fatalf("Failed to Initialize Exporter: %v", err)
	}

	tp := newTraceProvider(exp)
	defer func() { _ = tp.Shutdown(ctx) }()
	otel.SetTracerProvider(tp)
	tracer := tp.Tracer("report_service")

	reportStore := server.initReportStore(mongoClient, tracer)
	cassandraStore, err := store.New(Logger, tracer)
	if err != nil {
		log.Printf("Error in server cassandra store.New(): %s", err.Error())
		log.Fatal(err)
	}
	defer cassandraStore.CloseSession()

	cassandraStore.CreateTables()

	replyPublisher := server.initPublisher(server.config.CreateReportReplySubject)
	commandSubscriber := server.initSubscriber(server.config.CreateReportCommandSubject, QueueGroup)

	reportService := server.initReportService(cassandraStore, reportStore, tracer)
	reportHandler := server.initReportHandler(reportService, tracer)

	server.initCreateEventHandler(reportService, replyPublisher, commandSubscriber)

	server.start(reportHandler)

}

func (server *Server) initMongoClient() *mongo.Client {
	client, err := store.GetMongoClient(server.config.ReportDBHost, server.config.ReportDBPort)
	if err != nil {
		log.Printf("Error in server initMongoClient(): %s", err.Error())
		log.Fatal(err)
	}
	return client
}

func newExporter(address string) (*jaeger.Exporter, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(address)))
	if err != nil {
		log.Printf("Error in server newExporter(): %s", err.Error())
		return nil, err
	}
	return exp, nil
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("auth_service"),
		),
	)

	if err != nil {
		log.Printf("Error in server newTraceProvider(): %s", err.Error())
		panic(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}

func (server *Server) initReportStore(client *mongo.Client, tracer trace.Tracer) domain.ReportStore {
	store2 := store.NewReportMongoDBStore(client, tracer, Logger)
	return store2
}

func (server *Server) initReportService(eventStore domain.EventStore, reportStore domain.ReportStore, tracer trace.Tracer) *application.ReportService {
	return application.NewReportService(eventStore, reportStore, tracer, Logger)
}

func (server *Server) initReportHandler(service *application.ReportService, tracer trace.Tracer) *handlers.ReportHandler {
	return handlers.NewReportHandler(service, tracer, Logger)
}

func (server *Server) initPublisher(subject string) saga.Publisher {
	publisher, err := nats.NewNATSPublisher(
		server.config.NatsHost, server.config.NatsPort,
		server.config.NatsUser, server.config.NatsPass, subject)
	if err != nil {
		log.Printf("Error in server initPublisher(): %s", err.Error())
		log.Fatal(err)
	}
	return publisher
}

func (server *Server) initSubscriber(subject string, queueGroup string) saga.Subscriber {
	subscriber, err := nats.NewNATSSubscriber(
		server.config.NatsHost, server.config.NatsPort,
		server.config.NatsUser, server.config.NatsPass, subject, queueGroup)
	if err != nil {
		log.Printf("Error in server initSubscriber(): %s", err.Error())
		log.Fatal(err)
	}
	return subscriber
}

func (server *Server) initCreateEventHandler(reportService *application.ReportService, publisher saga.Publisher, subscriber saga.Subscriber) {
	_, err := handlers.NewCreateEventCommandHandler(reportService, publisher, subscriber)
	if err != nil {
		log.Printf("Error in server initCreateEventHandler(): %s", err.Error())
		log.Fatal(err)
	}
}

// start
func (server *Server) start(authHandler *handlers.ReportHandler) {
	router := mux.NewRouter()
	authHandler.Init(router)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", server.config.Port),
		Handler: router,
	}

	wait := time.Second * 15
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Error Shutting Down Server %s", err)
	}
	log.Println("Server Gracefully Stopped")
}

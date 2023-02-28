package startup

import (
	"context"
	"fmt"
	"follow_service/application"
	"follow_service/domain"
	"follow_service/handlers"
	"follow_service/startup/config"
	"follow_service/store"
	"github.com/gorilla/mux"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/sirupsen/logrus"
	saga "github.com/zjalicf/twitter-clone-common/common/saga/messaging"
	"github.com/zjalicf/twitter-clone-common/common/saga/messaging/nats"
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
	"syscall"
	"time"
)

type Server struct {
	config *config.Config
}

func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
	}
}

var Logger = logrus.New()

const (
	QueueGroup  = "follow_service"
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

func (server *Server) initNeo4JDriver() *neo4j.DriverWithContext {
	driver, err := store.GetClient(server.config.FollowDBHost, server.config.FollowDBPort,
		server.config.FollowDBUser, server.config.FollowDBPass)
	if err != nil {
		log.Fatal(err)
	}
	return driver
}

func (server *Server) Start() {

	initLogger()

	cfg := config.NewConfig()

	ctx := context.Background()
	exp, err := newExporter(cfg.JaegerAddress)
	if err != nil {
		log.Fatalf("Failed to Initialize Exporter: %v", err)
	}

	tp := newTraceProvider(exp)
	defer func() { _ = tp.Shutdown(ctx) }()
	otel.SetTracerProvider(tp)
	tracer := tp.Tracer("follow_service")

	neo4jDriver := server.initNeo4JDriver()
	followStore := server.initFollowStore(neo4jDriver, tracer, Logger)
	followService := server.initFollowService(followStore, tracer, Logger)
	followHandler := server.initFollowHandler(followService, tracer, Logger)

	//saga init
	replyPublisher := server.initPublisher(server.config.CreateUserReplySubject)
	commandSubscriber := server.initSubscriber(server.config.CreateUserCommandSubject, QueueGroup)

	server.initCreateUserHandler(followService, replyPublisher, commandSubscriber, tracer)

	server.start(followHandler)
}

func (server *Server) initFollowStore(driver *neo4j.DriverWithContext, tracer trace.Tracer, logging *logrus.Logger) domain.FollowRequestStore {
	store := store.NewFollowNeo4JStore(driver, tracer, logging)

	return store
}

func (server *Server) initFollowService(store domain.FollowRequestStore, tracer trace.Tracer, logging *logrus.Logger) *application.FollowService {
	return application.NewFollowService(store, tracer, logging)
}

func (server *Server) initFollowHandler(service *application.FollowService, tracer trace.Tracer, logging *logrus.Logger) *handlers.FollowHandler {
	return handlers.NewFollowHandler(service, tracer, logging)
}

func (server *Server) initCreateUserHandler(service *application.FollowService, publisher saga.Publisher, subscriber saga.Subscriber, tracer trace.Tracer) {
	_, err := handlers.NewCreateUserCommandHandler(service, publisher, subscriber, tracer)
	if err != nil {
		log.Fatal(err)
	}
}

func (server *Server) initPublisher(subject string) saga.Publisher {
	publisher, err := nats.NewNATSPublisher(
		server.config.NatsHost, server.config.NatsPort,
		server.config.NatsUser, server.config.NatsPass, subject)
	if err != nil {
		log.Fatal(err)
	}
	return publisher
}

func (server *Server) initSubscriber(subject string, queueGroup string) saga.Subscriber {
	subscriber, err := nats.NewNATSSubscriber(
		server.config.NatsHost, server.config.NatsPort,
		server.config.NatsUser, server.config.NatsPass, subject, queueGroup)
	if err != nil {
		log.Fatal(err)
	}
	return subscriber
}

func (server *Server) start(followHandler *handlers.FollowHandler) {
	router := mux.NewRouter()
	followHandler.Init(router)

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

func newExporter(address string) (*jaeger.Exporter, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(address)))
	if err != nil {
		return nil, err
	}
	return exp, nil
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("follow_service"),
		),
	)

	if err != nil {
		panic(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}

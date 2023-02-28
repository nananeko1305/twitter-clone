package startup

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
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
	"tweet_service/application"
	application2 "tweet_service/application"
	"tweet_service/domain"
	"tweet_service/handlers"
	"tweet_service/startup/config"
	"tweet_service/store"
	store2 "tweet_service/store"
)

var Logger = logrus.New()

type Server struct {
	config *config.Config
}

const (
	QueueGroup  = "tweet_service"
	LogFilePath = "/app/logs/application.log"
)

func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
	}
}

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
	tracer := tp.Tracer("tweet_service")

	redisClient := server.initRedisClient()
	tweetCache := server.initTweetCache(redisClient, tracer)
	tweetStore, err := store.New(log.Default(), tracer, Logger)
	if err != nil {
		log.Fatal(err)
	}
	defer tweetStore.CloseSession()
	tweetStore.CreateTables()

	//orchestrator
	commandPublisher := server.initPublisher(server.config.CreateReportCommandSubject)
	replySubscriber := server.initSubscriber(server.config.CreateReportReplySubject, QueueGroup)

	createReportOrchestrator := server.initCreateEventOrchestrator(commandPublisher, replySubscriber, tracer)

	tweetService := server.initTweetService(*tweetStore, tweetCache, tracer, createReportOrchestrator, Logger)

	tweetHandler := server.initTweetHandler(tweetService, tracer, Logger)

	server.start(tweetHandler)
}

func (server *Server) initTweetService(store store.TweetRepo, cache domain.TweetCache, tracer trace.Tracer, orchestrator *application2.CreateEventOrchestrator, logging *logrus.Logger) *application.TweetService {
	service := application.NewTweetService(&store, cache, tracer, orchestrator, Logger)
	Logger.Info("Started tweet service")
	return service
}

func (server *Server) initTweetCache(client *redis.Client, tracer trace.Tracer) domain.TweetCache {
	cache := store2.NewTweetRedisCache(client, tracer)
	return cache
}

func (server *Server) initTweetHandler(service *application.TweetService, tracer trace.Tracer, logging *logrus.Logger) *handlers.TweetHandler {
	return handlers.NewTweetHandler(service, tracer, logging)
}

func (server *Server) initRedisClient() *redis.Client {
	client, err := store2.GetRedisClient(server.config.TweetCacheHost, server.config.TweetCachePort)
	if err != nil {
		log.Fatal(err)
	}
	return client
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

func (server *Server) initCreateEventOrchestrator(publisher saga.Publisher, subscriber saga.Subscriber, tracer trace.Tracer) *application.CreateEventOrchestrator {
	orchestrator, err := application.NewCreateEventOrchestrator(publisher, subscriber, tracer)
	if err != nil {
		log.Fatalf("Error in server of report_service, line 182: %s", err.Error())
	}
	return orchestrator
}

func (server *Server) start(tweetHandler *handlers.TweetHandler) {
	router := mux.NewRouter()
	tweetHandler.Init(router)

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
			semconv.ServiceNameKey.String("tweet_service"),
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

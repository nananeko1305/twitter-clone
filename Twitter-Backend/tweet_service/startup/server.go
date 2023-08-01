package startup

import (
	"context"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	nats2 "github.com/nats-io/nats.go"
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
	"tweet_service/configs"
	"tweet_service/domain"
	"tweet_service/handlers"
	"tweet_service/startup/config"
	"tweet_service/store"
	store2 "tweet_service/store"
)

type Server struct {
	config *config.Config
}

const (
	QueueGroup = "tweet_service"
)

func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
	}
}

func (server *Server) Start() {

	cfg := config.NewConfig()

	client, elasticApi, err := configs.ConnectToElastic(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Elasticsearch: %v", err)
	}

	//connect to nats client
	natsConnection := configs.ConnectToNats(cfg)

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
	tweetStore, err := store.New(tracer)
	if err != nil {
		log.Fatal(err)
	}
	defer tweetStore.CloseSession()
	tweetStore.CreateTables()

	elasticStore := server.initTweetElastic(client, elasticApi)

	tweetService := server.initTweetService(*tweetStore, tweetCache, tracer, natsConnection, elasticStore)
	tweetService.SubscribeToNats(natsConnection)
	err = elasticStore.CheckIndex()
	if err != nil {
		log.Fatal(err)
	}

	tweetHandler := server.initTweetHandler(tweetService, tracer)

	server.start(tweetHandler)
}

func (server *Server) initTweetService(store store.TweetRepo, cache domain.TweetCache, tracer trace.Tracer, natsConnection *nats2.Conn, elasticService domain.TweetElasticStore) *application.TweetService {
	service := application.NewTweetService(&store, cache, tracer, natsConnection, elasticService)
	return service
}

func (server *Server) initTweetCache(client *redis.Client, tracer trace.Tracer) domain.TweetCache {
	cache := store2.NewTweetRedisCache(client, tracer)
	return cache
}

func (server *Server) initTweetElastic(client *elasticsearch.Client, elasticApi *esapi.API) domain.TweetElasticStore {
	cache := store2.NewTweetElasticStoreImpl(client, elasticApi)
	return cache
}

func (server *Server) initTweetHandler(service *application.TweetService, tracer trace.Tracer) *handlers.TweetHandler {
	return handlers.NewTweetHandler(service, tracer)
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

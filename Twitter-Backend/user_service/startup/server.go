package startup

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/olivere/elastic/v7"
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
	"syscall"
	"time"
	"user_service/application"
	"user_service/domain"
	"user_service/handlers"
	"user_service/startup/config"
	"user_service/store"
)

type Server struct {
	config *config.Config
}

const (
	QueueGroup = "user_service"
)

func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
	}
}

func (server *Server) Start() {

	client, err := elastic.NewClient(
		elastic.SetURL(server.config.ELASTICSEARCH_HOSTS),
	)
	if err != nil {
		return
	}

	mongoClient := server.initMongoClient()
	defer func(mongoClient *mongo.Client, ctx context.Context) {
		err := mongoClient.Disconnect(ctx)
		if err != nil {
			log.Println(err)
		}
	}(mongoClient, context.Background())

	//saga init
	replyPublisher := server.initPublisher(server.config.CreateUserReplySubject)
	commandSubscriber := server.initSubscriber(server.config.CreateUserCommandSubject, QueueGroup)

	cfg := config.NewConfig()

	ctx := context.Background()
	exp, err := newExporter(cfg.JaegerAddress)
	if err != nil {
		log.Fatalf("Failed to Initialize Exporter: %v", err)
	}

	tp := newTraceProvider(exp)
	defer func() { _ = tp.Shutdown(ctx) }()
	otel.SetTracerProvider(tp)
	tracer := tp.Tracer("user_service")

	userElasticStore := server.initUserElasticStore(client)
	userElasticStore.CheckIndex()
	userStore := server.initUserStore(mongoClient, tracer)
	userService := server.initUserService(userStore, userElasticStore, tracer)
	userHandler := server.initUserHandler(userService, tracer)

	server.initCreateUserHandler(userService, replyPublisher, commandSubscriber, tracer)

	server.start(userHandler)
}

func (server *Server) initUserStore(client *mongo.Client, tracer trace.Tracer) domain.UserStore {
	userStore := store.NewUserMongoDBStore(client, tracer)
	return userStore
}

func (server *Server) initUserElasticStore(client *elastic.Client) domain.UserElasticStore {
	userElasticStore := store.NewUserElasticStoreImpl(client)
	return userElasticStore
}

func (server *Server) initUserService(store domain.UserStore, olivereElastic domain.UserElasticStore, tracer trace.Tracer) *application.UserService {
	return application.NewUserService(store, olivereElastic, tracer)
}

func (server *Server) initUserHandler(service *application.UserService, tracer trace.Tracer) *handlers.UserHandler {
	return handlers.NewUserHandler(service, tracer)
}

func (server *Server) initMongoClient() *mongo.Client {
	client, err := store.GetClient(server.config.UserDBHost, server.config.UserDBPort)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func (server *Server) initCreateUserHandler(service *application.UserService, publisher saga.Publisher, subscriber saga.Subscriber, tracer trace.Tracer) {
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

func (server *Server) start(userHandler *handlers.UserHandler) {
	router := mux.NewRouter()
	userHandler.Init(router)

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
			semconv.ServiceNameKey.String("user_service"),
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

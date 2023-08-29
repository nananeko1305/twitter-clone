package startup

import (
	"auth_service/application"
	"auth_service/configs"
	"auth_service/domain"
	"auth_service/handlers"
	"auth_service/startup/config"
	store2 "auth_service/store"
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	nats2 "github.com/nats-io/nats.go"
	saga "github.com/zjalicf/twitter-clone-common/common/saga/messaging"
	"github.com/zjalicf/twitter-clone-common/common/saga/messaging/nats"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/option"
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

const (
	QueueGroup = "auth_service"
)

func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
	}
}

func (server *Server) Start() {

	firebaseClient, err := server.initFirebaseCloudMessaging()
	if err != nil {
		log.Println(err)
		return
	}

	mongoClient := server.initMongoClient()
	defer func(mongoClient *mongo.Client, ctx context.Context) {
		err := mongoClient.Disconnect(ctx)
		if err != nil {
			log.Println(err)
		}
	}(mongoClient, context.Background())

	natsConnection := configs.ConnectToNats(server.config)

	redisClient := server.initRedisClient()
	authCache := server.initAuthCache(redisClient)
	authStore := server.initAuthStore(mongoClient)

	//saga init

	//orchestrator
	commandPublisher := server.initPublisher(server.config.CreateUserCommandSubject)
	replySubscriber := server.initSubscriber(server.config.CreateUserReplySubject, QueueGroup)

	//service
	replyPublisher := server.initPublisher(server.config.CreateUserReplySubject)
	commandSubscriber := server.initSubscriber(server.config.CreateUserCommandSubject, QueueGroup)

	createUserOrchestrator := server.initCreateUserOrchestrator(commandPublisher, replySubscriber)

	authService := server.initAuthService(authStore, authCache, createUserOrchestrator, firebaseClient, natsConnection)
	authService.SubscribeToNats(natsConnection)
	server.initCreateUserHandler(authService, replyPublisher, commandSubscriber)
	authHandler := server.initAuthHandler(authService)

	server.start(authHandler)
}

func (server *Server) initMongoClient() *mongo.Client {
	client, err := store2.GetMongoClient(server.config.AuthDBHost, server.config.AuthDBPort)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func (server *Server) initRedisClient() *redis.Client {
	client, err := store2.GetRedisClient(server.config.AuthCacheHost, server.config.AuthCachePort)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func (server *Server) initAuthStore(client *mongo.Client) domain.AuthStore {
	store := store2.NewAuthMongoDBStore(client)
	return store
}

func (server *Server) initAuthCache(client *redis.Client) domain.AuthCache {
	cache := store2.NewAuthRedisCache(client)
	return cache
}

func (server *Server) initAuthService(store domain.AuthStore, cache domain.AuthCache, orchestrator *application.CreateUserOrchestrator, firebaseClient *messaging.Client, natsConnection *nats2.Conn) *application.AuthService {
	return application.NewAuthService(store, cache, orchestrator, firebaseClient, natsConnection)
}

func (server *Server) initAuthHandler(service *application.AuthService) *handlers.AuthHandler {
	return handlers.NewAuthHandler(service)
}

//saga

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

func (server *Server) initCreateUserOrchestrator(publisher saga.Publisher, subscriber saga.Subscriber) *application.CreateUserOrchestrator {
	orchestrator, err := application.NewCreateUserOrchestrator(publisher, subscriber)
	if err != nil {
		log.Fatal(err)
	}
	return orchestrator
}

func (server *Server) initCreateUserHandler(service *application.AuthService, publisher saga.Publisher, subscriber saga.Subscriber) {
	_, err := handlers.NewCreateUserCommandHandler(service, publisher, subscriber)
	if err != nil {
		log.Fatal(err)
	}
}

func (server *Server) initFirebaseCloudMessaging() (*messaging.Client, error) {
	opt := option.WithCredentialsFile("firebase_key.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error initializing messaging client: %v", err)
	}

	return client, nil
}

// start
func (server *Server) start(authHandler *handlers.AuthHandler) {
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

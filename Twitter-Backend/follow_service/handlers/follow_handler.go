package handlers

import (
	"encoding/json"
	"follow_service/application"
	"follow_service/authorization"
	"follow_service/domain"
	"follow_service/errors"
	"github.com/casbin/casbin"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
)

type FollowHandler struct {
	service            *application.FollowService
	counterUnavailable int
	tracer             trace.Tracer
}

func NewFollowHandler(service *application.FollowService, tracer trace.Tracer) *FollowHandler {
	return &FollowHandler{
		service:            service,
		counterUnavailable: 3,
		tracer:             tracer,
	}
}

func (handler *FollowHandler) Init(router *mux.Router) {

	authEnforcer, err := casbin.NewEnforcerSafe("./auth_model.conf", "./policy.csv")
	if err != nil {
		log.Fatal(err)
	}

	router.HandleFunc("/requests/", handler.GetRequestsForUser).Methods("GET")
	router.HandleFunc("/requests/{visibility}", handler.CreateRequest).Methods("POST")
	router.HandleFunc("/acceptRequest/{id}", handler.AcceptRequest).Methods("PUT")
	router.HandleFunc("/declineRequest/{id}", handler.DeclineRequest).Methods("PUT")
	router.HandleFunc("/feedInfo", handler.GetFeedInfoOfUser).Methods("GET")
	router.HandleFunc("/followings/{username}", handler.GetFollowingsOfUser).Methods("GET")
	router.HandleFunc("/followers/{username}", handler.GetFollowersOfUser).Methods("GET")
	router.HandleFunc("/followExist/{username}", handler.FollowExist).Methods("GET")
	router.HandleFunc("/recommendations", handler.GetRecommendationsForUser).Methods("GET")
	router.HandleFunc("/ad", handler.SaveAd).Methods("POST")

	http.Handle("/", router)
	log.Println("Successful")
	log.Fatal(http.ListenAndServe(":8004", authorization.Authorizer(authEnforcer)(router)))
}

func (handler *FollowHandler) GetRequestsForUser(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "FollowHandler.GetRequestsForUser")
	defer span.End()

	token, err := authorization.GetToken(req)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	claims := authorization.GetMapClaims(token.Bytes())

	returnRequests, err := handler.service.GetRequestsForUser(ctx, claims["username"])
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonResponse(returnRequests, writer)
}

func (handler *FollowHandler) GetFeedInfoOfUser(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "FollowHandler.GetFeedInfoOfUser")
	defer span.End()

	token, _ := authorization.GetToken(req)
	claims := authorization.GetMapClaims(token.Bytes())
	username := claims["username"]

	feedInfo, err := handler.service.GetFeedInfoOfUser(ctx, username)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(feedInfo, writer)

}

func (handler *FollowHandler) GetFollowingsOfUser(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "FollowHandler.GetFollowingsOfUser")
	defer span.End()

	vars := mux.Vars(req)
	var username string
	if vars["username"] == "me" {
		token, _ := authorization.GetToken(req)
		claims := authorization.GetMapClaims(token.Bytes())
		username = claims["username"]
	} else {
		username = vars["username"]
	}

	users, err := handler.service.GetFollowingsOfUser(ctx, username)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(users, writer)

}

func (handler *FollowHandler) GetFollowersOfUser(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "FollowHandler.GetFollowersOfUser")
	defer span.End()

	vars := mux.Vars(req)
	var username string
	if vars["username"] == "me" {
		token, _ := authorization.GetToken(req)
		claims := authorization.GetMapClaims(token.Bytes())
		username = claims["username"]
	} else {
		username = vars["username"]
	}

	users, err := handler.service.GetFollowersOfUser(ctx, username)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(users, writer)

}

func (handler *FollowHandler) GetRecommendationsForUser(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "FollowHandler.GetRecommendationsForUser")
	defer span.End()

	token, _ := authorization.GetToken(req)
	claims := authorization.GetMapClaims(token.Bytes())
	username := claims["username"]

	users, err := handler.service.GetRecommendationsByUsername(ctx, username)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(users, writer)
}

func (handler *FollowHandler) CreateRequest(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "FollowHandler.CreateRequest")
	defer span.End()

	var request domain.FollowRequest
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Header.Get("Authorization") == "" {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := authorization.GetToken(req)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusUnauthorized)
	}
	claims := authorization.GetMapClaims(token.Bytes())

	vars := mux.Vars(req)
	var visibility bool
	if vars["visibility"] == "private" {
		visibility = true
	} else {
		visibility = false
	}

	err = handler.service.CreateRequest(ctx, &request, claims["username"], visibility)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (handler *FollowHandler) AcceptRequest(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "FollowHandler.AcceptRequest")
	defer span.End()

	vars := mux.Vars(req)
	followId, ok := vars["id"]
	if !ok {
		http.Error(writer, errors.BadRequestError, http.StatusBadRequest)
	}

	err := handler.service.AcceptRequest(ctx, &followId)
	if err != nil {
		http.Error(writer, errors.BadRequestError, http.StatusBadRequest)
	}

	writer.WriteHeader(http.StatusOK)

}

func (handler *FollowHandler) DeclineRequest(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "FollowHandler.GetAll")
	defer span.End()

	vars := mux.Vars(req)
	followId, ok := vars["id"]
	if !ok {
		http.Error(writer, errors.BadRequestError, http.StatusBadRequest)
	}

	err := handler.service.DeclineRequest(ctx, &followId)
	if err != nil {
		http.Error(writer, errors.BadRequestError, http.StatusBadRequest)
	}

	writer.WriteHeader(http.StatusOK)

}

func (handler *FollowHandler) SaveAd(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "FollowHandler.SaveAd")
	defer span.End()

	var request domain.Ad
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		http.Error(writer, "bad request", http.StatusBadRequest)
		return
	}

	err = handler.service.SaveAd(ctx, &request)
	if err != nil {
		http.Error(writer, "internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (handler *FollowHandler) FollowExist(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "FollowHandler.FollowExist")
	defer span.End()

	vars := mux.Vars(req)
	followingUsername, ok := vars["username"]
	if !ok {
		http.Error(writer, errors.BadRequestError, http.StatusBadRequest)
		return
	}

	token, _ := authorization.GetToken(req)
	claims := authorization.GetMapClaims(token.Bytes())
	username := claims["username"]

	request := domain.FollowRequest{
		Receiver:  followingUsername,
		Requester: username,
	}

	isExist, err := handler.service.FollowExist(ctx, &request)
	if err != nil {
		http.Error(writer, errors.InternalServerError, http.StatusInternalServerError)
		return
	}

	jsonResponse(isExist, writer)
}

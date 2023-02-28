package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"tweet_service/application"
	"tweet_service/authorization"
	"tweet_service/domain"

	"github.com/casbin/casbin"
	"github.com/cristalhq/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type TweetHandler struct {
	service *application.TweetService
	tracer  trace.Tracer
	logging *logrus.Logger
}

var jwtKey = []byte(os.Getenv("SECRET_KEY"))
var verifier, _ = jwt.NewVerifierHS(jwt.HS256, jwtKey)

func NewTweetHandler(service *application.TweetService, tracer trace.Tracer,  logging *logrus.Logger) *TweetHandler {
	return &TweetHandler{
		service: service,
		tracer:  tracer,
		logging: logging,
	}
}

func (handler *TweetHandler) Init(router *mux.Router) {
	authEnforcer, err := casbin.NewEnforcerSafe("./auth_model.conf", "./policy.csv")
	log.Println("successful init of enforcer")
	if err != nil {
		log.Fatal(err)
	}

	router.Use(ExtractTraceInfoMiddleware)
	router.HandleFunc("/", handler.GetAll).Methods("GET")
	router.HandleFunc("/getOneTweet/{id}", handler.GetOne).Methods("GET")
	router.HandleFunc("/", Post(handler)).Methods("POST")
	router.HandleFunc("/", handler.GetAll).Methods("GET")
	router.HandleFunc("/image/{id}", handler.GetTweetImage).Methods("GET")
	router.HandleFunc("/favorite", handler.Favorite).Methods("POST")
	router.HandleFunc("/user/{username}", handler.GetTweetsByUser).Methods("GET")
	router.HandleFunc("/whoLiked/{id}", handler.GetLikesByTweet).Methods("GET")
	router.HandleFunc("/feed", handler.GetFeedByUser).Methods("GET")
	router.HandleFunc("/retweet", handler.Retweet).Methods("POST")
	router.HandleFunc("/timespent", handler.TimespentOnAd).Methods("POST")
	router.HandleFunc("/viewCount", handler.ViewProfileFromAdd).Methods("POST")

	http.Handle("/", router)
	log.Println("Successful")
	log.Fatal(http.ListenAndServe(":8001", authorization.Authorizer(authEnforcer)(router)))
}

func (handler *TweetHandler) GetAll(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "TweetHandler.GetAll")
	defer span.End()

	handler.logging.Infoln("tweetHandler.GetAll : getAll endpoint reached")

	tweets, err := handler.service.GetAll(ctx)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonResponse(tweets, writer)
}

func (handler *TweetHandler) GetOne(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "TweetHandler.GetOne")
	defer span.End()

	handler.logging.Infoln("tweetHandler.Get : get endpoint reached")

	vars := mux.Vars(req)
	tweetID := vars["id"]

	tweets, err := handler.service.GetOne(ctx, tweetID)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonResponse(tweets, writer)
}

func (handler *TweetHandler) GetTweetsByUser(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "TweetHandler.GetTweetsByUser")
	defer span.End()

	handler.logging.Infoln("tweetHandler.TweetsByUser : TweetsByUser endpoint reached")

	vars := mux.Vars(req)
	username, ok := vars["username"]
	if !ok {
		handler.logging.Errorln("tweetHandler.TweetsByUser : bad username")
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	tweets, err := handler.service.GetTweetsByUser(ctx, username)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonResponse(tweets, writer)
}

func (handler *TweetHandler) GetLikesByTweet(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "TweetHandler.GetLikesByTweet")
	defer span.End()

	handler.logging.Infoln("tweetHandler.likesByTweet : likesByTweet endpoint reached")

	vars := mux.Vars(req)
	tweetID, ok := vars["id"]
	if !ok {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	favorites, err := handler.service.GetLikesByTweet(ctx, tweetID)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonResponse(favorites, writer)
}

func (handler *TweetHandler) Favorite(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "TweetHandler.Favorite")
	defer span.End()

	handler.logging.Infoln("tweetHandler.favorite : Favorite endpoint reached")

	bearer := req.Header.Get("Authorization")
	bearerToken := strings.Split(bearer, "Bearer ")
	tokenString := bearerToken[1]

	token, err := jwt.Parse([]byte(tokenString), verifier)

	if err != nil {
		log.Println(err)
		http.Error(writer, "unauthorized", http.StatusUnauthorized)
		return
	}

	claims := authorization.GetMapClaims(token.Bytes())
	username := claims["username"]

	var tweet domain.Tweet
	err = json.NewDecoder(req.Body).Decode(&tweet)
	if err != nil {
		handler.logging.Errorln("tweetHandler.favorite : tweet bad body")
		log.Println(err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	tweets, err := handler.service.Favorite(ctx, tweet.ID.String(), username, tweet.Advertisement)

	if err != nil {
		log.Printf("Error in tweetHandler Favorite(): %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	jsonResponse(tweets, writer)

}

func (handler *TweetHandler) Post(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "TweetHandler.Post")
	defer span.End()

	handler.logging.Infoln("tweetHandler.Post : post endpoint reached")

	err := req.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Println(err)
		return
	}

	strs := strings.Split(req.Header.Get("Content-Type"), "; boundary")

	if strs[0] != "multipart/form-data" {
		log.Println("Invalid Content-Type")
		handler.logging.Infoln("tweetHandler.post : image error")
		http.Error(writer, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	file, _, err := req.FormFile("image")

	var imageBytes []byte
	if err == nil {
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Fprintln(writer, "Error reading image:", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		imageBytes = bytes
		defer file.Close()
	}

	tweet := req.FormValue("json")

	//json without AdConfig
	var tweetVal domain.Tweet
	err = json.Unmarshal([]byte(tweet), &tweetVal)
	if err != nil {
		handler.logging.Errorln("tweetHandler.post- unmarshal error")
		log.Printf("Error in TweetHandler.Post unmarshal json 1")
	}

	//json with AdConfig
	var tweetAdVal domain.AdTweet
	err = json.Unmarshal([]byte(tweet), &tweetAdVal)
	if err != nil {
		handler.logging.Errorln("tweetHandler.post- unmarshal error")
		log.Printf("Error in TweetHandler.Post unmarshal json 2")
	}

	if err != nil {
		handler.logging.Errorln("tweetHandler.post- unmarshal error")
		http.Error(writer, "bad json format", http.StatusBadRequest)

	}

	//sta dalje?

	if req.Header.Get("Authorization") == "" {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	bearerToken := strings.Split(req.Header.Get("Authorization"), "Bearer ")
	tokenString := bearerToken[1]
	token := authorization.GetToken(tokenString)

	claims := authorization.GetMapClaims(token.Bytes())
	username := claims["username"]

	//dodati adve
	ret, err := handler.service.Post(ctx, &tweetVal, username, &imageBytes)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	writer.WriteHeader(http.StatusOK)
	jsonResponse(ret, writer)
}

func (handler *TweetHandler) GetTweetImage(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "TweetHandler.GetTweetImage")
	defer span.End()

	handler.logging.Infoln("tweetHandler.getTweetImage reached")

	vars := mux.Vars(req)
	id, ok := vars["id"]
	if !ok {
		http.Error(writer, "bad request", http.StatusBadRequest)
		return
	}

	image, err := handler.service.GetTweetImage(ctx, id)
	if err != nil {
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(*image)
}

func Post(handler *TweetHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler.Post(w, r)
	}
}

func ExtractTraceInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (handler *TweetHandler) GetFeedByUser(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "TweetHandler.GetFeedByUser")
	defer span.End()

	handler.logging.Infoln("tweetHandler.getFeedByUser reached")

	feed, err := handler.service.GetFeedByUser(ctx, req.Header.Get("Authorization"))
	if err != nil {
		log.Printf("error: %s", err.Error())
		if err.Error() == "FollowServiceError" {
			http.Error(writer, err.Error(), http.StatusServiceUnavailable)
			return
		}
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	} else {
		jsonResponse(feed, writer)

	}
}

func (handler *TweetHandler) Retweet(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "TweetHandler.Retweet")
	defer span.End()

	handler.logging.Infoln("tweetHandler.retweet reached")

	bearer := req.Header.Get("Authorization")
	bearerToken := strings.Split(bearer, "Bearer ")
	tokenString := bearerToken[1]

	token, err := jwt.Parse([]byte(tokenString), verifier)

	if err != nil {
		log.Println(err)
		http.Error(writer, "unauthorized", http.StatusUnauthorized)
		return
	}

	claims := authorization.GetMapClaims(token.Bytes())
	username := claims["username"]

	var tweetID domain.TweetID
	err = json.NewDecoder(req.Body).Decode(&tweetID)
	if err != nil {
		log.Printf("Error in decoding http request body in TweetHandler.Retweet: %s", err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	code, err := handler.service.Retweet(ctx, tweetID.ID, username)
	if err != nil {
		http.Error(writer, err.Error(), code)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (handler *TweetHandler) TimespentOnAd(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "TweetHandler.TimespentOnAd")
	defer span.End()

	handler.logging.Infoln("tweetHandler.timeSpentOnAd reached")

	var timespent domain.Timespent
	err := json.NewDecoder(req.Body).Decode(&timespent)
	if err != nil {
		log.Println("Error in decoding body in handler function TimespentOnAd")
		http.Error(writer, "bad request", http.StatusBadRequest)
	}

	handler.service.TimeSpentOnAd(ctx, &timespent)

	writer.WriteHeader(http.StatusOK)
}

func (handler *TweetHandler) ViewProfileFromAdd(writer http.ResponseWriter, req *http.Request) {
	ctx, span := handler.tracer.Start(req.Context(), "TweetHandler.ViewProfileFromAdd")
	defer span.End()

	handler.logging.Infoln("tweetHandler.ViewProfile reached")

	var tweetID domain.TweetID
	err := json.NewDecoder(req.Body).Decode(&tweetID)
	if err != nil {
		log.Println("Error in decoding body in handler function ViewProfileFromAdd")
		http.Error(writer, "bad request", http.StatusBadRequest)
	}

	handler.service.ViewProfileFromAd(ctx, tweetID)

	writer.WriteHeader(http.StatusOK)
}

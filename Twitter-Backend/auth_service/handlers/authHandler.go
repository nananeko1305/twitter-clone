package handlers

import (
	"auth_service/application"
	"auth_service/authorization"
	"auth_service/domain"
	"auth_service/errors"
	"auth_service/store"
	"context"
	"encoding/json"
	"fmt"
	"github.com/casbin/casbin"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"io"
	"log"
	"net/http"
	"strings"
)

type KeyUser struct{}

type AuthHandler struct {
	service *application.AuthService
	store   *store.AuthMongoDBStore
}

func NewAuthHandler(service *application.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (handler *AuthHandler) Init(router *mux.Router) {
	authEnforcer, err := casbin.NewEnforcerSafe("./auth_model.conf", "./policy.csv")
	if err != nil {
		log.Fatal(err)
	}

	router.Use(ExtractTraceInfoMiddleware)
	loginRouter := router.Methods(http.MethodPost).Subrouter()
	loginRouter.HandleFunc("/login", handler.Login)

	registerRouter := router.Methods(http.MethodPost).Subrouter()
	registerRouter.HandleFunc("/register", handler.Register)
	registerRouter.Use(MiddlewareUserValidation)

	verifyRouter := router.Methods(http.MethodPost).Subrouter()
	verifyRouter.HandleFunc("/verifyAccount", handler.VerifyAccount)

	router.HandleFunc("/login", handler.Login).Methods("POST")
	router.HandleFunc("/register", handler.Register).Methods("POST")
	router.HandleFunc("/verifyAccount", handler.VerifyAccount).Methods("POST")
	router.HandleFunc("/resendVerify", handler.ResendVerificationToken).Methods("POST")
	router.HandleFunc("/recoverPasswordToken", handler.SendRecoveryPasswordToken).Methods("POST")
	router.HandleFunc("/checkRecoverToken", handler.CheckRecoveryPasswordToken).Methods("POST")
	router.HandleFunc("/recoverPassword", handler.RecoverPassword).Methods("POST")
	router.HandleFunc("/changePassword", handler.ChangePassword).Methods("POST")
	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":8003", authorization.Authorizer(authEnforcer)(router)))

}

func (handler *AuthHandler) Register(writer http.ResponseWriter, req *http.Request) {

	myUser := req.Context().Value(domain.User{}).(domain.User)

	token, statusCode, err := handler.service.Register(context.Background(), &myUser)
	if statusCode == 55 {
		writer.WriteHeader(http.StatusFound)
		http.Error(writer, err.Error(), 302)
		return
	}
	if err != nil {
		http.Error(writer, err.Error(), statusCode)
		return
	}
	writer.WriteHeader(http.StatusOK)
	jsonResponse(token, writer)
}

func (handler *AuthHandler) VerifyAccount(writer http.ResponseWriter, req *http.Request) {

	var request domain.RegisterRecoverVerification
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		log.Println(err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if len(request.UserToken) == 0 {
		http.Error(writer, errors.InvalidUserTokenError, http.StatusBadRequest)
		return
	}

	err = handler.service.VerifyAccount(context.Background(), &request)
	if err != nil {
		if err.Error() == errors.InvalidTokenError {
			log.Println(err.Error())
			http.Error(writer, errors.InvalidTokenError, http.StatusNotAcceptable)
		} else if err.Error() == errors.ExpiredTokenError {
			log.Println(err.Error())
			http.Error(writer, errors.ExpiredTokenError, http.StatusNotFound)
		}
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (handler *AuthHandler) ResendVerificationToken(writer http.ResponseWriter, req *http.Request) {

	var request domain.ResendVerificationRequest
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		http.Error(writer, errors.InvalidRequestFormatError, http.StatusBadRequest)
		log.Fatal(err.Error())
		return
	}

	err = handler.service.ResendVerificationToken(context.Background(), &request)
	if err != nil {
		if err.Error() == errors.InvalidResendMailError {
			http.Error(writer, err.Error(), http.StatusNotAcceptable)
			return
		} else {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writer.WriteHeader(http.StatusOK)
}

func (handler *AuthHandler) SendRecoveryPasswordToken(writer http.ResponseWriter, req *http.Request) {

	buf := new(strings.Builder)
	_, err := io.Copy(buf, req.Body)
	if err != nil {
		http.Error(writer, errors.InvalidRequestFormatError, http.StatusBadRequest)
		log.Fatal(err.Error())
		return
	}

	id, statusCode, err := handler.service.SendRecoveryPasswordToken(context.Background(), buf.String())
	if err != nil {
		http.Error(writer, err.Error(), statusCode)
		return
	}

	jsonResponse(id, writer)
}

func (handler *AuthHandler) CheckRecoveryPasswordToken(writer http.ResponseWriter, req *http.Request) {

	var request domain.RegisterRecoverVerification
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		http.Error(writer, errors.InvalidRequestFormatError, http.StatusBadRequest)
		log.Fatal(err.Error())
		return
	}

	err = handler.service.CheckRecoveryPasswordToken(context.Background(), &request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusNotAcceptable)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (handler *AuthHandler) RecoverPassword(writer http.ResponseWriter, req *http.Request) {

	var request domain.RecoverPasswordRequest
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		http.Error(writer, errors.InvalidRequestFormatError, http.StatusBadRequest)
		log.Fatal(err.Error())
		return
	}

	err = handler.service.RecoverPassword(context.Background(), &request)
	if err != nil {
		if err.Error() == errors.NotMatchingPasswordsError {
			http.Error(writer, err.Error(), http.StatusNotAcceptable)
			return
		}
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (handler *AuthHandler) Login(writer http.ResponseWriter, req *http.Request) {

	var request domain.Credentials
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := handler.service.Login(context.Background(), &request)
	if err != nil {
		if err.Error() == errors.NotVerificatedUser {
			http.Error(writer, token, http.StatusLocked)
			return
		}
		http.Error(writer, "Username not exist!", http.StatusBadRequest)
		return
	}

	if token == "not_same" {
		http.Error(writer, "Wrong password", http.StatusUnauthorized)
		return
	}

	writer.Write([]byte(token))
}

func MiddlewareUserValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {

		user := &domain.User{}
		err := user.FromJSON(request.Body)
		if err != nil {
			http.Error(responseWriter, "Unable to Decode JSON", http.StatusBadRequest)
			return
		}

		err = user.ValidateUser()
		if err != nil {
			http.Error(responseWriter, fmt.Sprintf("Validation Error:\n %s.", err), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(request.Context(), domain.User{}, *user)
		request = request.WithContext(ctx)

		next.ServeHTTP(responseWriter, request)
	})
}

func (handler *AuthHandler) ChangePassword(writer http.ResponseWriter, req *http.Request) {

	var token string = req.Header.Get("Authorization")
	bearerToken := strings.Split(token, "Bearer ")
	tokenString := bearerToken[1]

	var password domain.PasswordChange
	err := json.NewDecoder(req.Body).Decode(&password)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	status := handler.service.ChangePassword(context.Background(), password, tokenString)

	if status == "oldPassErr" {
		http.Error(writer, "Wrong old password", http.StatusConflict) //409
		return
	} else if status == "newPassErr" {
		http.Error(writer, "Wrong new password", http.StatusNotAcceptable) //406
		return
	} else if status == "baseErr" {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func ExtractTraceInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

package application

import (
	"auth_service/authorization"
	"auth_service/domain"
	"auth_service/errors"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cristalhq/jwt/v4"
	"github.com/google/uuid"
	"github.com/zjalicf/twitter-clone-common/common/saga/create_user"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	userServiceHost = os.Getenv("USER_SERVICE_HOST")
	userServicePort = os.Getenv("USER_SERVICE_PORT")
	smtpServer      = "smtp-mail.outlook.com"
	smtpServerPort  = 587
	smtpEmail       = os.Getenv("SMTP_AUTH_MAIL")
	smtpPassword    = os.Getenv("SMTP_AUTH_PASSWORD")
)

type AuthService struct {
	store        domain.AuthStore
	cache        domain.AuthCache
	orchestrator *CreateUserOrchestrator
}

func NewAuthService(store domain.AuthStore, cache domain.AuthCache, orchestrator *CreateUserOrchestrator) *AuthService {
	return &AuthService{
		store:        store,
		cache:        cache,
		orchestrator: orchestrator,
	}
}

func (service *AuthService) GetAll(ctx context.Context) ([]*domain.Credentials, error) {

	return service.store.GetAll(ctx)
}

func (service *AuthService) Register(ctx context.Context, user *domain.User) (string, int, error) {

	isUsernameExists, err := checkBlackList(user.Password)

	if isUsernameExists {
		return "", 55, fmt.Errorf("Password not acceptable, try another one!")
	}

	_, err = service.store.GetOneUser(ctx, user.Username)
	if err == nil {
		return "", 406, fmt.Errorf(errors.UsernameAlreadyExist)
	}

	userServiceEndpointMail := fmt.Sprintf("http://%s:%s/mailExist/%s", userServiceHost, userServicePort, user.Email)
	userServiceRequestMail, _ := http.NewRequest("GET", userServiceEndpointMail, nil)
	response, err := http.DefaultClient.Do(userServiceRequestMail)
	if err != nil {
		return "", 500, fmt.Errorf(errors.ServiceUnavailable)
	}
	if response.StatusCode != 404 {
		return "", 406, fmt.Errorf(errors.EmailAlreadyExist)
	}

	//provereni su mejl i username

	user.ID = primitive.NewObjectID()
	validatedUser, err := validateUserType(user)
	if err != nil {
		return "", 0, err
	}

	pass := []byte(user.Password)
	hash, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)

	if err != nil {
		return "", 500, err
	}
	user.Password = string(hash)

	credentials := domain.Credentials{
		ID:       user.ID,
		Username: validatedUser.Username,
		Password: user.Password,
		UserType: validatedUser.UserType,
		Verified: false,
	}

	err = service.store.Register(ctx, &credentials)
	if err != nil {
		return "", 0, err
	}

	err = service.orchestrator.Start(ctx, validatedUser)
	if err != nil {
		return "", 0, err
	}

	return credentials.ID.Hex(), 200, nil
}

func (service *AuthService) DeleteUserByID(ctx context.Context, id primitive.ObjectID) error {

	return service.store.DeleteUserByID(ctx, id)
}

func (service *AuthService) SendMail(ctx context.Context, user *domain.User) error {

	validationToken := uuid.New()
	err := service.cache.PostCacheData(user.ID.Hex(), validationToken.String())
	if err != nil {
		log.Fatalf("failed to post validation data to redis: %s", err)
		return err
	}

	err = service.sendValidationMail(ctx, validationToken, user.Email)
	if err != nil {
		log.Printf("Failed to send mail: %s", err.Error())
		return err
	}

	return nil
}

func (service *AuthService) sendValidationMail(ctx context.Context, validationToken uuid.UUID, email string) error {

	message := gomail.NewMessage()
	message.SetHeader("From", smtpEmail)
	message.SetHeader("To", email)
	message.SetHeader("Subject", "Verify your Twitter Clone account")

	bodyString := fmt.Sprintf("Your validation token for twitter account is:\n%s", validationToken)
	message.SetBody("text", bodyString)

	client := gomail.NewDialer(smtpServer, smtpServerPort, smtpEmail, smtpPassword)

	_, err := client.Dial()
	if err != nil {
		return err
	}

	if err := client.DialAndSend(message); err != nil {
		log.Fatalf("failed to send verification mail because of: %s", err)
		return err
	}

	return nil
}

func (service *AuthService) VerifyAccount(ctx context.Context, validation *domain.RegisterRecoverVerification) error {

	token, err := service.cache.GetCachedValue(validation.UserToken)
	if err != nil {
		log.Println(errors.ExpiredTokenError)
		return fmt.Errorf(errors.ExpiredTokenError)
	}

	if validation.MailToken == token {
		err = service.cache.DelCachedValue(validation.UserToken)
		if err != nil {
			return err
		}

		userID, err := primitive.ObjectIDFromHex(validation.UserToken)
		user := service.store.GetOneUserByID(ctx, userID)
		user.Verified = true

		err = service.store.UpdateUser(ctx, user)
		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf(errors.InvalidTokenError)
}

func (service *AuthService) ResendVerificationToken(ctx context.Context, request *domain.ResendVerificationRequest) error {

	if len(request.UserMail) == 0 {
		return fmt.Errorf(errors.InvalidResendMailError)
	}

	tokenUUID, _ := uuid.NewUUID()

	err := service.cache.PostCacheData(request.UserToken, tokenUUID.String())
	if err != nil {
		return err
	}

	err = service.sendValidationMail(ctx, tokenUUID, request.UserMail)
	if err != nil {
		return err
	}

	return nil
}

func (service *AuthService) SendRecoveryPasswordToken(ctx context.Context, email string) (string, int, error) {

	userServiceEndpoint := fmt.Sprintf("http://%s:%s/mailExist/%s", userServiceHost, userServicePort, email)
	userServiceRequest, _ := http.NewRequest("GET", userServiceEndpoint, nil)
	response, _ := http.DefaultClient.Do(userServiceRequest)
	if response.StatusCode != 200 {
		if response.StatusCode == 404 {
			return "", 404, fmt.Errorf(errors.NotFoundMailError)
		}
	}

	buf := new(strings.Builder)
	_, _ = io.Copy(buf, response.Body)
	userID := buf.String()

	recoverUUID, _ := uuid.NewUUID()
	err := service.sendRecoverPasswordMail(ctx, recoverUUID, email)
	if err != nil {
		return "", 500, err
	}

	err = service.cache.PostCacheData(userID, recoverUUID.String())
	if err != nil {
		return "", 500, err
	}

	return userID, 200, nil
}

func (service *AuthService) CheckRecoveryPasswordToken(ctx context.Context, request *domain.RegisterRecoverVerification) error {

	if len(request.UserToken) == 0 {
		return fmt.Errorf(errors.InvalidUserTokenError)
	}

	token, err := service.cache.GetCachedValue(request.UserToken)
	if err != nil {
		return fmt.Errorf(errors.InvalidTokenError)
	}

	if request.MailToken != token {
		return fmt.Errorf(errors.InvalidTokenError)
	}

	_ = service.cache.DelCachedValue(request.UserToken)
	return nil
}

func (service *AuthService) sendRecoverPasswordMail(ctx context.Context, validationToken uuid.UUID, email string) error {

	message := gomail.NewMessage()
	message.SetHeader("From", smtpEmail)
	message.SetHeader("To", email)
	message.SetHeader("Subject", "Recover password on your Twitter Clone account")

	bodyString := fmt.Sprintf("Your recover password token is:\n%s", validationToken)
	message.SetBody("text", bodyString)

	client := gomail.NewDialer(smtpServer, smtpServerPort, smtpEmail, smtpPassword)

	if err := client.DialAndSend(message); err != nil {
		log.Fatalf("AuthService.sendRecoverPasswordMail : failed to send verification mail because of: %s", err)
		return err
	}

	return nil
}

func (service *AuthService) RecoverPassword(ctx context.Context, recoverPassword *domain.RecoverPasswordRequest) error {

	if recoverPassword.NewPassword != recoverPassword.RepeatedNew {
		return fmt.Errorf(errors.NotMatchingPasswordsError)
	}

	primitiveID, err := primitive.ObjectIDFromHex(recoverPassword.UserID)
	if err != nil {
		return err
	}
	credentials := service.store.GetOneUserByID(ctx, primitiveID)

	pass := []byte(recoverPassword.NewPassword)
	hash, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	credentials.Password = string(hash)

	err = service.store.UpdateUser(ctx, credentials)
	if err != nil {
		return err
	}

	return nil
}

func (service *AuthService) Login(ctx context.Context, credentials *domain.Credentials) (string, error) {

	user, err := service.store.GetOneUser(ctx, credentials.Username)
	if err != nil {
		return "", err
	}

	if !user.Verified {
		userServiceEndpoint := fmt.Sprintf("http://%s:%s/%s", userServiceHost, userServicePort, user.ID.Hex())
		userServiceRequest, _ := http.NewRequest("GET", userServiceEndpoint, nil)
		response, _ := http.DefaultClient.Do(userServiceRequest)
		if response.StatusCode != 200 {
			if response.StatusCode == 404 {
				return "", fmt.Errorf("user doesn't exist")
			}
		}

		var userUser domain.User
		err := responseToType(response.Body, &userUser)
		if err != nil {
			return "", err
		}

		verify := domain.ResendVerificationRequest{
			UserToken: user.ID.Hex(),
			UserMail:  userUser.Email,
		}

		err = service.ResendVerificationToken(ctx, &verify)
		if err != nil {
			return "", err
		}

		return user.ID.Hex(), fmt.Errorf(errors.NotVerificatedUser)
	}

	passError := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if passError != nil {
		return "not_same", err
	}

	tokenString, err := GenerateJWT(user)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func responseToType(response io.ReadCloser, user *domain.User) error {
	responseBodyBytes, err := io.ReadAll(response)
	if err != nil {
		log.Printf("err in readAll %s", err.Error())
		return err
	}

	err = json.Unmarshal(responseBodyBytes, &user)
	if err != nil {
		log.Printf("err in Unmarshal %s", err.Error())
		return err
	}

	return nil
}

func GenerateJWT(user *domain.Credentials) (string, error) {
	key := []byte(os.Getenv("SECRET_KEY"))
	signer, err := jwt.NewSignerHS(jwt.HS256, key)
	if err != nil {
		log.Println(err)
	}

	builder := jwt.NewBuilder(signer)

	claims := &domain.Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.UserType,
		ExpiresAt: time.Now().Add(time.Minute * 60),
	}

	token, err := builder.Build(claims)
	if err != nil {
		log.Println(err)
	}

	return token.String(), nil
}

func (service *AuthService) ChangePassword(ctx context.Context, password domain.PasswordChange, token string) string {

	parsedToken := authorization.GetToken(token)
	claims := authorization.GetMapClaims(parsedToken.Bytes())

	username := claims["username"]

	user, err := service.store.GetOneUser(ctx, username)
	if err != nil {
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password.OldPassword))
	if err != nil {
		return "oldPassErr"
	}

	var validNew bool = false
	if password.NewPassword == password.NewPasswordConfirm {
		validNew = true
	}

	if validNew {
		newEncryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Println(err)
			return "hashErr"
		}

		user.Password = string(newEncryptedPassword)

		err = service.store.UpdateUser(ctx, user)
		if err != nil {
			return "baseErr"
		}

	} else {
		return "newPassErr"

	}

	return "ok"
}

func (service *AuthService) UserToDomain(userIn create_user.User) domain.User {
	var user domain.User
	user.ID = userIn.ID
	user.Firstname = userIn.Firstname
	user.Lastname = userIn.Lastname
	if userIn.Gender == "Male" {
		user.Gender = "Male"
	} else {
		user.Gender = "Female"
	}
	user.Age = userIn.Age
	user.Residence = userIn.Residence
	user.Email = userIn.Email
	user.Username = userIn.Username
	user.Password = userIn.Password
	if userIn.UserType == "Regular" {
		user.UserType = "Regular"
	} else {
		user.UserType = "Business"
	}
	user.Visibility = userIn.Visibility
	user.CompanyName = userIn.CompanyName
	user.Website = userIn.Website

	return user
}

func (service *AuthService) DomainToUser(userIn *domain.User) create_user.User {
	var user create_user.User
	user.ID = userIn.ID
	user.Firstname = userIn.Firstname
	user.Lastname = userIn.Lastname
	if userIn.Gender == "Male" {
		user.Gender = "Male"
	} else {
		user.Gender = "Female"
	}
	user.Age = userIn.Age
	user.Residence = userIn.Residence
	user.Email = userIn.Email
	user.Username = userIn.Username
	user.Password = userIn.Password
	if userIn.UserType == "Regular" {
		user.UserType = "Regular"
	} else {
		user.UserType = "Business"
	}
	user.Visibility = userIn.Visibility
	user.CompanyName = userIn.CompanyName
	user.Website = userIn.Website

	return user
}

func validateUserType(user *domain.User) (*domain.User, error) {

	if user.UserType == domain.Admin {
		user.UserType = domain.Admin
		return user, nil
	}

	business := isBusiness(user)
	regular := isRegular(user)

	if business && regular {
		return nil, fmt.Errorf("invalid user format")
	} else if business {
		user.UserType = domain.Business
		return user, nil
	} else if regular {
		user.UserType = domain.Regular
		return user, nil
	}

	return nil, fmt.Errorf("invalid user data")
}

func isBusiness(user *domain.User) bool {
	if len(user.CompanyName) >= 3 &&
		len(user.Website) >= 3 &&
		len(user.Email) >= 3 &&
		len(user.Username) >= 3 {
		return true
	}

	return false
}

func isRegular(user *domain.User) bool {
	if len(user.Firstname) >= 3 &&
		len(user.Lastname) >= 3 &&
		len(user.Gender) >= 3 &&
		user.Age >= 1 &&
		len(user.Residence) >= 3 &&
		len(user.Username) >= 3 {
		return true
	}

	return false
}

func checkBlackList(username string) (bool, error) {

	file, err := os.Open("blacklist.txt")
	if err != nil {
		log.Printf("Error in authService.checkBlackList: %s", err.Error())
		return false, err
	}
	defer file.Close()

	blacklist := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		blacklist[scanner.Text()] = true
	}
	if blacklist[username] {
		return true, nil
	} else {
		return false, nil
	}
}

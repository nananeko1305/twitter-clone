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
	"github.com/sirupsen/logrus"
	"github.com/zjalicf/twitter-clone-common/common/saga/create_user"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
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
	tracer       trace.Tracer
	orchestrator *CreateUserOrchestrator
	logging      *logrus.Logger
}

func NewAuthService(store domain.AuthStore, cache domain.AuthCache, orchestrator *CreateUserOrchestrator, tracer trace.Tracer, logging *logrus.Logger) *AuthService {
	return &AuthService{
		store:        store,
		cache:        cache,
		orchestrator: orchestrator,
		tracer:       tracer,
		logging:      logging,
	}
}

func (service *AuthService) GetAll(ctx context.Context) ([]*domain.Credentials, error) {
	ctx, span := service.tracer.Start(ctx, "AuthService.GetAll")
	defer span.End()

	service.logging.Infoln("AuthService.GetAll : getAll service reached")

	return service.store.GetAll(ctx)
}

func (service *AuthService) Register(ctx context.Context, user *domain.User) (string, int, error) {
	ctx, span := service.tracer.Start(ctx, "AuthService.Register")
	defer span.End()
	service.logging.Infoln("AuthService.Register : register service reached")

	isUsernameExists, err := checkBlackList(user.Password)
	log.Println(isUsernameExists)

	if isUsernameExists {
		service.logging.Errorln("AuthService.Register : password is in blacklist")
		return "", 55, fmt.Errorf("Password not acceptable, try another one!")
	}

	_, err = service.store.GetOneUser(ctx, user.Username)
	if err == nil {
		service.logging.Errorln("AuthService.Register : username exists in database")
		return "", 406, fmt.Errorf(errors.UsernameAlreadyExist)
	}

	userServiceEndpointMail := fmt.Sprintf("http://%s:%s/mailExist/%s", userServiceHost, userServicePort, user.Email)
	userServiceRequestMail, _ := http.NewRequest("GET", userServiceEndpointMail, nil)
	response, err := http.DefaultClient.Do(userServiceRequestMail)
	if err != nil {
		service.logging.Errorf("AuthService.Register : %s (user_service unavailable)", err)
		return "", 500, fmt.Errorf(errors.ServiceUnavailable)
	}
	if response.StatusCode != 404 {
		service.logging.Errorln("AuthService.Register : email exists in database")
		return "", 406, fmt.Errorf(errors.EmailAlreadyExist)
	}

	//provereni su mejl i username

	user.ID = primitive.NewObjectID()
	validatedUser, err := validateUserType(user)
	if err != nil {
		service.logging.Errorf("AuthService.Register : %s", err)
		service.logging.Errorln(err)
		return "", 0, err
	}

	pass := []byte(user.Password)
	hash, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)

	service.logging.Infoln("AuthService.Register : password sucessfully hashed")

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
		service.logging.Errorf("AuthService.Register : %s", err)
		return "", 0, err
	}

	err = service.orchestrator.Start(ctx, validatedUser)
	if err != nil {
		service.logging.Errorln("AuthService.Register : orchestrator error")
		return "", 0, err
	}

	service.logging.Infoln("AuthService.Register : register service finished")
	return credentials.ID.Hex(), 200, nil
}

func (service *AuthService) DeleteUserByID(ctx context.Context, id primitive.ObjectID) error {
	ctx, span := service.tracer.Start(ctx, "AuthService.DeleteUserByID")
	defer span.End()
	service.logging.Infoln("AuthService.DeleteUserByID : deleteUser service reached")

	return service.store.DeleteUserByID(ctx, id)
}

func (service *AuthService) SendMail(ctx context.Context, user *domain.User) error {
	ctx, span := service.tracer.Start(ctx, "AuthService.SendMail")
	defer span.End()
	service.logging.Infoln("AuthService.SendMail : sendMail service reached")

	validationToken := uuid.New()
	err := service.cache.PostCacheData(user.ID.Hex(), validationToken.String())
	if err != nil {
		service.logging.Errorf("AuthService.SendMail : failed to post validation data to redis: %s \n", err)
		log.Fatalf("failed to post validation data to redis: %s", err)
		return err
	}

	err = service.sendValidationMail(ctx, validationToken, user.Email)
	if err != nil {
		service.logging.Errorf("AuthService.SendMail : %s (failed to sent mail)", err)
		log.Printf("Failed to send mail: %s", err.Error())
		return err
	}
	service.logging.Infoln("AuthService.SendMail : sendMail service finished")

	return nil
}

func (service *AuthService) sendValidationMail(ctx context.Context, validationToken uuid.UUID, email string) error {
	ctx, span := service.tracer.Start(ctx, "AuthService.sendValidationMail")
	defer span.End()
	service.logging.Infoln("AuthService.sendValidationMail : sendValidationMail service reached")

	message := gomail.NewMessage()
	message.SetHeader("From", smtpEmail)
	message.SetHeader("To", email)
	message.SetHeader("Subject", "Verify your Twitter Clone account")

	bodyString := fmt.Sprintf("Your validation token for twitter account is:\n%s", validationToken)
	message.SetBody("text", bodyString)

	client := gomail.NewDialer(smtpServer, smtpServerPort, smtpEmail, smtpPassword)

	_, err := client.Dial()
	if err != nil {
		service.logging.Errorf("AuthService.sendValidationMail : %s", err)
		return err
	}

	if err := client.DialAndSend(message); err != nil {
		service.logging.Errorln("AuthService.sendValidationMail : failed to send verification mail because of: %s", err)
		log.Fatalf("failed to send verification mail because of: %s", err)
		return err
	}

	service.logging.Infoln("AuthService.sendValidationMail : sendValidationMail service finished")

	return nil
}

func (service *AuthService) VerifyAccount(ctx context.Context, validation *domain.RegisterRecoverVerification) error {
	ctx, span := service.tracer.Start(ctx, "AuthService.VerifyAccount")
	defer span.End()
	service.logging.Infoln("AuthService.VerifyAccount : verifyAccount service reached")

	token, err := service.cache.GetCachedValue(validation.UserToken)
	if err != nil {
		log.Println(errors.ExpiredTokenError)
		return fmt.Errorf(errors.ExpiredTokenError)
	}

	if validation.MailToken == token {
		err = service.cache.DelCachedValue(validation.UserToken)
		if err != nil {
			service.logging.Errorf("AuthService.VerifyAccount : error in deleting cached value: %s", err)
			return err
		}

		userID, err := primitive.ObjectIDFromHex(validation.UserToken)
		user := service.store.GetOneUserByID(ctx, userID)
		user.Verified = true

		err = service.store.UpdateUser(ctx, user)
		if err != nil {
			service.logging.Errorf("AuthService.VerifyAccount : error in updating user after changing status of verify: %s", err)
			return err
		}

		return nil
	}

	return fmt.Errorf(errors.InvalidTokenError)
}

func (service *AuthService) ResendVerificationToken(ctx context.Context, request *domain.ResendVerificationRequest) error {
	ctx, span := service.tracer.Start(ctx, "AuthService.ResendVerificationToken")
	defer span.End()
	service.logging.Infoln("AuthService.ResendVerificationToken : resendVerification service reached")

	if len(request.UserMail) == 0 {
		service.logging.Errorf("AuthService.ResendVerificationToken : %s", errors.InvalidResendMailError)
		return fmt.Errorf(errors.InvalidResendMailError)
	}

	tokenUUID, _ := uuid.NewUUID()

	err := service.cache.PostCacheData(request.UserToken, tokenUUID.String())
	if err != nil {
		service.logging.Errorf("AuthService.ResendVerificationToken.PostCacheData() : %s", err)
		return err
	}

	err = service.sendValidationMail(ctx, tokenUUID, request.UserMail)
	if err != nil {
		service.logging.Errorf("AuthService.ResendVerificationToken.sendValidationMail() : %s", err)
		return err
	}

	return nil
}

func (service *AuthService) SendRecoveryPasswordToken(ctx context.Context, email string) (string, int, error) {
	ctx, span := service.tracer.Start(ctx, "AuthService.SendRecoveryPasswordToken")
	defer span.End()
	service.logging.Infoln("AuthService.SendRecoveryPasswordToken : sendRecoveryPass service reached")

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
		service.logging.Errorf("AuthService.SendRecoveryPasswordToken.sendRecoverPasswordMail() : %s", err)
		return "", 500, err
	}

	err = service.cache.PostCacheData(userID, recoverUUID.String())
	if err != nil {
		service.logging.Errorf("AuthService.SendRecoveryPasswordToken.PostCacheData() : %s", err)
		return "", 500, err
	}

	return userID, 200, nil
}

func (service *AuthService) CheckRecoveryPasswordToken(ctx context.Context, request *domain.RegisterRecoverVerification) error {
	ctx, span := service.tracer.Start(ctx, "AuthService.CheckRecoveryPasswordToken")
	defer span.End()

	service.logging.Infoln("AuthService.CheckRecoveryPasswordToken : checkRecovery service reached")

	if len(request.UserToken) == 0 {
		service.logging.Errorln("AuthService.CheckRecoveryPasswordToken : invalid user token")
		return fmt.Errorf(errors.InvalidUserTokenError)
	}

	token, err := service.cache.GetCachedValue(request.UserToken)
	if err != nil {
		service.logging.Errorf("AuthService.CheckRecoveryPasswordToken.GetCachedValue() : %s", err)
		return fmt.Errorf(errors.InvalidTokenError)
	}

	if request.MailToken != token {
		service.logging.Errorf("AuthService.CheckRecoveryPasswordToken : %s", err)
		return fmt.Errorf(errors.InvalidTokenError)
	}

	_ = service.cache.DelCachedValue(request.UserToken)
	return nil
}

func (service *AuthService) sendRecoverPasswordMail(ctx context.Context, validationToken uuid.UUID, email string) error {
	ctx, span := service.tracer.Start(ctx, "AuthService.sendRecoverPasswordMail")
	defer span.End()

	service.logging.Infoln("AuthService.sendRecoverPasswordMail : sendRecoveryMail service reached")

	message := gomail.NewMessage()
	message.SetHeader("From", smtpEmail)
	message.SetHeader("To", email)
	message.SetHeader("Subject", "Recover password on your Twitter Clone account")

	bodyString := fmt.Sprintf("Your recover password token is:\n%s", validationToken)
	message.SetBody("text", bodyString)

	client := gomail.NewDialer(smtpServer, smtpServerPort, smtpEmail, smtpPassword)

	if err := client.DialAndSend(message); err != nil {
		service.logging.Errorln(err)
		log.Fatalf("AuthService.sendRecoverPasswordMail : failed to send verification mail because of: %s", err)
		return err
	}

	return nil
}

func (service *AuthService) RecoverPassword(ctx context.Context, recoverPassword *domain.RecoverPasswordRequest) error {
	ctx, span := service.tracer.Start(ctx, "AuthService.RecoverPassword")
	defer span.End()
	service.logging.Infoln("AuthService.RecoverPassword : recoverPassword service reached")

	if recoverPassword.NewPassword != recoverPassword.RepeatedNew {
		service.logging.Errorln("AuthService.RecoverPassword : password don't match")
		return fmt.Errorf(errors.NotMatchingPasswordsError)
	}

	primitiveID, err := primitive.ObjectIDFromHex(recoverPassword.UserID)
	if err != nil {
		service.logging.Errorf("AuthService.RecoverPassword.ObjectIDFromHex() : %s", err)
		return err
	}
	credentials := service.store.GetOneUserByID(ctx, primitiveID)

	pass := []byte(recoverPassword.NewPassword)
	hash, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	if err != nil {
		service.logging.Errorf("AuthService.RecoverPassword.GenerateFromPassword() : %s", err)
		return err
	}
	credentials.Password = string(hash)

	err = service.store.UpdateUser(ctx, credentials)
	if err != nil {
		service.logging.Errorf("AuthService.RecoverPassword.UpdateUser() : %s", err)
		return err
	}

	return nil
}

func (service *AuthService) Login(ctx context.Context, credentials *domain.Credentials) (string, error) {
	ctx, span := service.tracer.Start(ctx, "AuthService.Login")
	defer span.End()
	service.logging.Infoln("AuthService.Login : login service reached")

	user, err := service.store.GetOneUser(ctx, credentials.Username)
	if err != nil {
		service.logging.Errorf("AuthService.Login.GetOneUser() %s", err)
		return "", err
	}

	if !user.Verified {
		userServiceEndpoint := fmt.Sprintf("http://%s:%s/%s", userServiceHost, userServicePort, user.ID.Hex())
		userServiceRequest, _ := http.NewRequest("GET", userServiceEndpoint, nil)
		response, _ := http.DefaultClient.Do(userServiceRequest)
		if response.StatusCode != 200 {
			if response.StatusCode == 404 {
				service.logging.Errorln("AuthService.Login : user doesn't exist")
				return "", fmt.Errorf("user doesn't exist")
			}
		}

		var userUser domain.User
		err := responseToType(response.Body, &userUser)
		if err != nil {
			service.logging.Errorf("AuthService.Login.responseToType() : %s", err)
			return "", err
		}

		verify := domain.ResendVerificationRequest{
			UserToken: user.ID.Hex(),
			UserMail:  userUser.Email,
		}

		err = service.ResendVerificationToken(ctx, &verify)
		if err != nil {
			service.logging.Errorf("AuthService.Login.ResendVerificationToken() : %s", err)
			return "", err
		}

		return user.ID.Hex(), fmt.Errorf(errors.NotVerificatedUser)
	}

	passError := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if passError != nil {
		service.logging.Errorln("AuthService.Login.CompareHashAndPassword() : password not same")
		return "not_same", err
	}

	tokenString, err := GenerateJWT(user)
	if err != nil {
		service.logging.Errorf("AuthService.Login.GenerateJWT() : %s", err)
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
	ctx, span := service.tracer.Start(ctx, "AuthService.ChangePassword")
	defer span.End()

	service.logging.Infoln("changePassword service reached")

	parsedToken := authorization.GetToken(token)
	claims := authorization.GetMapClaims(parsedToken.Bytes())

	username := claims["username"]

	user, err := service.store.GetOneUser(ctx, username)
	if err != nil {
		service.logging.Errorf("AuthService.ChangePassword.GetOneUser() : %s", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password.OldPassword))
	if err != nil {
		service.logging.Errorf("AuthService.ChangePassword.CompareHashAndPassword() : %s (old password not match)", err)
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
			service.logging.Errorf("AuthService.ChangePassword.UpdateUser() : %s", err)
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

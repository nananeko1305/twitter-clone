package domain

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"regexp"
	"time"
)

type User struct {
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	Firstname  string             `bson:"firstName,omitempty" json:"firstName,omitempty" validate:"onlyChar"`
	Lastname   string             `bson:"lastName,omitempty" json:"lastName,omitempty" validate:"onlyChar"`
	Gender     Gender             `bson:"gender,omitempty" json:"gender,omitempty" validate:"onlyChar"`
	Age        int                `bson:"age,omitempty" json:"age,omitempty"`
	Residence  string             `bson:"residence,omitempty" json:"residence,omitempty" validate:"onlyCharAndSpace"`
	Email      string             `bson:"email" json:"email" validate:"required,email"`
	Username   string             `bson:"username" json:"username" validate:"onlyCharAndNum,required"`
	Password   string             `bson:"password" json:"password" validate:"onlyCharAndNum,required"`
	UserType   UserType           `bson:"userType" json:"userType" validate:"onlyChar"`
	Visibility bool               `bson:"visibility" json:"visibility"`

	CompanyName string `bson:"companyName,omitempty" json:"companyName,omitempty" validation:"onlyCharAndNum"`
	Website     string `bson:"website,omitempty" json:"website,omitempty" validate:"onlyCharAndNum"`
}

type Gender string

const (
	Male   = "Male"
	Female = "Female"
)

type UserType string

const (
	Regular  = "Regular"
	Business = "Business"
)

type Credentials struct {
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	Username string             `bson:"username" json:"username"`
	Password string             `bson:"password" json:"password"`
	UserType UserType           `bson:"userType" json:"userType"`
	Verified bool               `bson:"verified" json:"verified"`
}

type PasswordChange struct {
	OldPassword        string `json:"old_password"`
	NewPassword        string `json:"new_password"`
	NewPasswordConfirm string `json:"new_password_confirm"`
}

type Claims struct {
	UserID    primitive.ObjectID `json:"user_id"`
	Username  string             `json:"username"`
	Role      UserType           `json:"userType"`
	ExpiresAt time.Time          `json:"expires_at"`
}

type RegisterRecoverVerification struct {
	UserToken string `json:"user_token"`
	MailToken string `json:"mail_token"`
}

type ResendVerificationRequest struct {
	UserToken string `json:"user_token"`
	UserMail  string `json:"user_mail"`
}

type ResetPasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
	RepeatedNew string `json:"repeated_new"`
}

type RecoverPasswordRequest struct {
	UserID      string `json:"id"`
	NewPassword string `json:"new_password"`
	RepeatedNew string `json:"repeated_new"`
}

func (user *User) ValidateUser() error {
	validate := validator.New()

	err := validate.RegisterValidation("onlyCharAndSpace", onlyCharactersAndSpaceField)
	if err != nil {
		return err
	}

	err = validate.RegisterValidation("onlyChar", onlyCharactersField)
	if err != nil {
		return err
	}

	err = validate.RegisterValidation("onlyCharAndNum", onlyCharactersAndNumbersField)
	if err != nil {
		return err
	}

	return validate.Struct(user)
}

// Allows only letters [a-z]
func onlyCharactersField(fl validator.FieldLevel) bool {
	re := regexp.MustCompile("[-_a-zA-Z]*")
	matches := re.FindAllString(fl.Field().String(), -1)

	if len(matches) != 1 {
		return false
	}

	return true
}

func onlyCharactersAndSpaceField(fl validator.FieldLevel) bool {
	re := regexp.MustCompile("[a-zA-Z ]*")
	matches := re.FindAllString(fl.Field().String(), -1)

	if len(matches) != 1 {
		return false
	}

	return true
}

// Allows only letters [a-z] and numbers [0-9]
func onlyCharactersAndNumbersField(fl validator.FieldLevel) bool {
	re := regexp.MustCompile("[-_a-zA-Z0-9]*")
	matches := re.FindAllString(fl.Field().String(), -1)

	if len(matches) != 1 {
		return false
	}

	return true
}

func (user *User) FromJSON(reader io.Reader) error {
	d := json.NewDecoder(reader)
	return d.Decode(user)
}

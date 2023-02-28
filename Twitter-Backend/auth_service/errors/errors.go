package errors

const (
	InvalidRequestFormatError = "invalid request format"
	InvalidTokenError         = "token is invalid"
	InvalidUserTokenError     = "invalid user token"
	ExpiredTokenError         = "verification token has expired"
	InvalidResendMailError    = "invalid resend mail"
	NotFoundMailError         = "mail not found in database"
	NotMatchingPasswordsError = "passwords doesn't matching"
	NotVerificatedUser        = "user wasn't verified yet"
	UsernameAlreadyExist      = "username already exists in database"
	EmailAlreadyExist         = "email already exists in database"
	ServiceUnavailable        = "service is unavailable at the moment"
)

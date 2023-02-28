package errors

var (
	ServiceUnavailable    = "service is unavailable at this moment, try again later"
	BadRequestError       = "bad request"
	ErrorInSaveFollow     = "save follow relationship error"
	ErrorInAcceptRequest  = "update accept request error"
	ErrorRequestNotExists = "request not exists"
	InternalServerError   = "internal server error"
)

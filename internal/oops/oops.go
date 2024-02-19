package oops

import "errors"

var (
	ErrOrderCreate           = errors.New("new order number accepted for processing")
	ErrOrderReady            = errors.New("the order number has already been uploaded by another user")
	ErrOrderNumberInvalid    = errors.New("invalid order number")
	ErrInsufficientFunds     = errors.New("insufficient funds")
	ErrEmptyData             = errors.New("no result")
	ErrLuhnValidate          = errors.New("invalid order format")
	ErrStatusNotOK           = errors.New("status not ok")
	ErrStatusTooManyRequests = errors.New("status too many requests")
	ErrInvalidToken          = errors.New("token invalid")
)

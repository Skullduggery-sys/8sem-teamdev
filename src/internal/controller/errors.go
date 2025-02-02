package controller

import "errors"

var (
	errInternal            = errors.New("internal error's occurred")
	errNotFound            = errors.New("not found")
	errUserNotFound        = errors.New("user not found")
	errInvalidArguments    = errors.New("requests' arguments are invalid")
	errActionNotAuthorized = errors.New("action not authorized")
)

package service

import "errors"

var (
	ErrNotFound              = errors.New("not found")
	ErrAdminIsNotAuthtorized = errors.New("user is authorized as an admin")
	ErrGeneratingHash        = errors.New("failed to generate hash from password")
	ErrBadPassword           = errors.New("password is not matched with the login")
	ErrLoginAlreadyExists    = errors.New("user with such login already exists")
	ErrBadVeryfyCode         = errors.New("verification code is not matched")
	ErrUserMetaNotFound      = errors.New("verification metadata not found")
	ErrVerifyActionNotSaved  = errors.New("verification action is lost")
	ErrWaiting2FA            = errors.New("wainting to finish 2FA")
)

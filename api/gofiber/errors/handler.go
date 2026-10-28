package errors

import "errors"

var (
	Invalidated = errors.New("validation error")

	ErrClientVersionTooOld            = errors.New("app version too old")
	ErrClientVersionInvalid           = errors.New("app version invalid")
	PasswordTooShort                  = errors.New("password too short")
	PasswordNotSet                    = errors.New("password not set")
	MaxCountReached                   = errors.New("max count reached")
	BodySizeLimitExceeded             = errors.New("body limit exceeded")
	Unexpected                        = errors.New("unexpected errors happened")
	ErrOnlyAvailableForSocialAccounts = errors.New("operation only available to social accounts")
	ErrOnlyAvailableForUserPassword   = errors.New("operation only available for account/password")
)

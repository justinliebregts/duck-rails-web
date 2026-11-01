package errors

import "errors"

var (
	ErrDuckzDuplicate             = errors.New("duckz duplicate key")
	ErrDuckzInvalidRequestID      = errors.New("duckz invalid request_id")
	ErrDuckzInvalidBars           = errors.New("duckz invalid bar count")
	ErrDuckzDepositCapReached     = errors.New("duckz hourly coin cap reached")
	ErrDuckzUnknownItem           = errors.New("duckz unknown store item")
	ErrDuckzInsufficientCoins     = errors.New("duckz insufficient coins")
	ErrDuckzUnknownSku            = errors.New("duckz unknown iap sku")
	ErrDuckzInvalidPlatformTxnID  = errors.New("duckz invalid platform_txn_id")
	ErrDuckzEntitlementNotOwned   = errors.New("duckz entitlement not verified with platform")
	ErrDuckzPlatformTxnIDUsed     = errors.New("duckz platform_txn_id already used")
	ErrDuckzUnsupportedPlatform   = errors.New("duckz iap platform not supported")
	ErrDuckzRequestIDOtherPlayer  = errors.New("duckz request_id belongs to another player")
	ErrDuckzPlayerNotFound        = errors.New("duckz player not found")
	ErrDuckzInvalidAmount         = errors.New("duckz invalid coin amount")
	ErrDuckzEquipmentNotSupported = errors.New("duckz has no equipment; only gem grants are supported")
	ErrDuckzItemNotOwned          = errors.New("duckz item not owned")
	ErrDuckzItemNotForSale        = errors.New("duckz item is only unlocked by an iap sku")
)

package oculus

import (
	"time"
)

const (
	UserNonceValidateUrl   = "/user_nonce_validate"
	VerifyItemOwnershipUrl = "/verify_entitlement"
	ConsumeIAPItemUrl      = "/consume_entitlement"
	RetrieveItemsOwnedUrl  = "/viewer_purchases"
)

var (
	requestTimeout = 5 * time.Second
)

//
//func DebugOculus() fiber.Map {
//	m := config.DebugOculus()
//	m["oculus_config"] = oculusConfig
//	m["request_timeout"] = requestTimeout
//	return m
//}

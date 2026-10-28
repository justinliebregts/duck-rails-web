package service

import (
	"gofiber/config"
)

// GetBanReasonByHash returns the standard reason map.
// It stays in the Service layer because it "knows" about the config.
func GetBanReasonByHash() map[string]string {
	//	BanReasonByHashVersionCheck = map[string]string{"by": "HashVersionCheckAndReport", "ver": config.RetrieveApiVersion()}

	return map[string]string{
		"by":  "HashVersionCheckAndReport",
		"ver": config.RetrieveApiVersion(),
	}
}

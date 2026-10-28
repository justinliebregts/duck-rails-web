package remote_auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

const (
	OculusAttestationEnableKey = "ENABLE_OCULUS_ATTESTATION"
)

var (
// OculusCredential OculusCredentials
//
//	oculusConfig = OculusConfig{
//		AppID:               "AP_ID",
//		PlatformServer:      "https://graph.oculus.com",
//		PlatformAccessToken: "OC|AP_ID|AP_SECRET",
//	}
)

type OculusCredentials struct {
	AppID     string
	AppSecret string
	//ApiBase   string
}

func (c *OculusCredentials) FormPlatformAccessToken() string {
	return fmt.Sprintf("OC|%v|%v", c.AppID, c.AppSecret)
}

func (c *OculusCredentials) GenerateSHA256SignatureWithOculusSecret(devPayload string) string {
	// Create a new HMAC using SHA256
	h := hmac.New(sha256.New, []byte(c.AppSecret))

	// Write the payload to the HMAC
	h.Write([]byte(devPayload))

	// Compute the HMAC signature
	signature := h.Sum(nil)

	// Return the signature as a hexadecimal string
	return "sha256=" + hex.EncodeToString(signature)
}

func loadOculusCredentials() OculusCredentials {
	return OculusCredentials{
		AppID:     os.Getenv("OCULUS_APPID"),
		AppSecret: os.Getenv("OCULUS_APPSECRET"),
		//ApiBase:   os.Getenv("OCULUS_API_BASE"),
	}
}

type OculusConfig struct {
	OculusAttestationEnabled bool
	OcCred                   OculusCredentials
	PlatformServer           string
	PlatformAccessToken      string
}

func (c *OculusConfig) DebugPrint() fiber.Map {
	return fiber.Map{
		"credential":   c.OcCred,
		"server_url":   c.PlatformServer,
		"access_token": c.PlatformAccessToken,
	}
}

func LoadOculusConfig() *OculusConfig {
	ocCred := loadOculusCredentials()

	return &OculusConfig{
		OculusAttestationEnabled: IsOculusAttestationEnabled(),
		OcCred:                   ocCred,
		PlatformServer:           "https://graph.oculus.com", // Default or from Env
		PlatformAccessToken:      ocCred.FormPlatformAccessToken(),
	}
}

func IsOculusAttestationEnabled() bool {
	if res, err := strconv.ParseBool(os.Getenv(OculusAttestationEnableKey)); err == nil {
		return res
	}

	//Default is FALSE
	return false
}

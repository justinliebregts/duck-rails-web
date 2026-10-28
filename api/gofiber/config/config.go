package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

var (
	subenv string

	fiberAppCustomErrorHandler fiber.ErrorHandler

	FiberAppConfig = fiber.Config{
		//Views: loadTemplateEngine(), // pure JSON at this moment
		ErrorHandler:      fiberAppCustomErrorHandler,
		EnablePrintRoutes: IsEnvDevelopment(),

		EnableTrustedProxyCheck: true,
		ProxyHeader:             fiber.HeaderXForwardedFor,
		// Add your proxy IPs or Docker subnets so Fiber trusts them:
		TrustedProxies: []string{
			"127.0.0.1",
			"172.16.0.0/12", // Covers Docker network bridge subnets (e.g., 172.19.0.x)
			"10.0.0.0/8",
			"192.168.0.0/16",
		},
	}
)

func SetCustomErrorHandler(hand fiber.ErrorHandler) {
	fiberAppCustomErrorHandler = hand
}

type BaseConfiguration struct {
	IsEnvDevelopment bool
	//ApiVersion       string
}

func (b *BaseConfiguration) Load() {
	b.IsEnvDevelopment = IsEnvDevelopment()
	//b.ApiVersion = RetrieveApiVersion()
}

// RetrieveApiVersion read config; give api version; v23...v99
func RetrieveApiVersion() string {
	return os.Getenv("API_VERSION")
}

func IsFakeLocalK8S() bool {
	return os.Getenv("FAKE_LOCAL_K8S") == "true"
}

// RetrieveAppVersion read config; give app version; [major].[minor].[patch]
func RetrieveAppVersion(platform string) string {
	switch platform {
	case "STEAM", "STEAM_PC":
		return os.Getenv("APP_STEAM_VERSION")
	case "STEAM_VR":
		return os.Getenv("APP_STEAM_VR_VERSION")
	case "GOOGLE", "ANDROID":
		return os.Getenv("APP_GOOGLE_VERSION")
	case "IOS", "APPLE":
		return os.Getenv("APP_IOS_VERSION")
	case "PICO":
		return os.Getenv("APP_PICO_VERSION")
	case "IQIYI":
		return os.Getenv("APP_IQIYI_VERSION")
	case "YVR":
		return os.Getenv("APP_YVR_VERSION")
	case "OCULUS":
		return os.Getenv("APP_VERSION")
	}

	return "UnKnOwN"
}

// func RetrieveAgoraAppID() string {
// 	return os.Getenv("AGORA_VOICE_APP_ID")
// }

// func RetrieveAgoraCertificate() string {
// 	return os.Getenv("AGORA_VOICE_APP_CERTIFICATE")
// }

func IsMonitorEnabled() bool {
	return os.Getenv("APP_MONITOR") == "true"
}

func IsRequestIDEnabled() bool {
	return os.Getenv("ENABLE_REQUEST_ID") == "true"
}

func IsEnvDevelopment() bool {
	return strings.HasPrefix(strings.ToLower(subenv), "dev")
}

func IsPhotonSessionCheckEnabled() bool {
	return os.Getenv("ENABLE_PHOTON_AUTH") == "1"
}

func IsPhotonSecretPhaseMatched(sec string) bool {
	return os.Getenv("PHOTON_AUTH_PHASE") == sec
}

//func DebugLogPrintln(v ...any) {
//	if IsEnvDevelopment() {
//		log.Println(v)
//	}
//}

func DebugLogPrintf(format string, v ...any) {
	if IsEnvDevelopment() {
		log.Printf(format, v...)
	}
}

func LoadDotEnv(filebase string) {
	subenv = os.Getenv("DEPLOY_ENV")
	if "" == subenv {
		subenv = "development"
	}

	log.Println("subenv:" + subenv)

	fileNames := []string{
		".env." + subenv + ".local",
		".env.local",
		".env." + subenv,
		".env",
	}

	_ = godotenv.Load(filepath.Join(filebase, fileNames[0]))
	if "test" != subenv {
		_ = godotenv.Load(filepath.Join(filebase, fileNames[1]))
	}

	_ = godotenv.Load(filepath.Join(filebase, fileNames[2]))
	_ = godotenv.Load(filepath.Join(filebase, fileNames[3]))
	_ = godotenv.Load() // The Original .env
	return
}

func RetrieveOSEnvironment() []string {
	return os.Environ()
}

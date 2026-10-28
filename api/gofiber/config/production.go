package config

import (
	"fmt"
	"gofiber/internal/rediskeys"
	"os"
	"strings"
)

// ValidateProductionConfiguration reports missing key names, never secret values.
func ValidateProductionConfiguration() error {
	if IsEnvDevelopment() || os.Getenv("DEPLOY_ENV") == "test" {
		return nil
	}
	required := []string{"REDIS_NAMESPACE", "API_VERSION", "DB_HOST", "DB_NAME", "DB_ACCOUNT", "DB_PASSWORD", "REDIS_HOST", "REDIS_PORT", "OCULUS_APPID", "OCULUS_APPSECRET"}
	if IsPhotonSessionCheckEnabled() {
		required = append(required, "PHOTON_AUTH_PHASE")
	}
	var missing []string
	for _, key := range required {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" || strings.HasPrefix(value, "REPLACE_") {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("configure these settings before startup: %s", strings.Join(missing, ", "))
	}
	return rediskeys.Validate(os.Getenv("REDIS_NAMESPACE"))
}

package config

import (
	"strings"
	"testing"
)

func TestProductionConfiguration(t *testing.T) {
	previous := subenv
	subenv = "production"
	t.Cleanup(func() { subenv = previous })
	t.Setenv("DEPLOY_ENV", "production")
	t.Setenv("REDIS_NAMESPACE", "grc")
	for _, key := range []string{"API_VERSION", "DB_HOST", "DB_NAME", "DB_ACCOUNT", "DB_PASSWORD", "REDIS_HOST", "REDIS_PORT", "OCULUS_APPID", "OCULUS_APPSECRET"} {
		t.Setenv(key, "configured")
	}
	t.Setenv("ENABLE_PHOTON_AUTH", "0")
	t.Setenv("PHOTON_AUTH_PHASE", "")
	if err := ValidateProductionConfiguration(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DB_HOST", "REPLACE_MYSQL_PRIVATE_HOST:3306")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("ENABLE_PHOTON_AUTH", "1")
	err := ValidateProductionConfiguration()
	if err == nil {
		t.Fatal("expected missing configuration error")
	}
	for _, key := range []string{"DB_HOST", "DB_PASSWORD", "PHOTON_AUTH_PHASE"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("missing key %s in %v", key, err)
		}
	}
	if strings.Contains(err.Error(), "configured") || strings.Contains(err.Error(), "REPLACE_MYSQL_PRIVATE_HOST") {
		t.Fatal("validation must not expose values")
	}
}

func TestJWTMissingOrMalformedKeyDoesNotPanic(t *testing.T) {
	t.Setenv("JWT_PRIVATE_KEY", "")
	for _, prefix := range []string{"", "no-comma", "begin,end"} {
		t.Setenv("JWT_PRIVATE_PREFIX_SUFFIX", prefix)
		if _, err := SetupRSAPrivateKey(); err == nil {
			t.Fatalf("expected error for invalid key with prefix %q", prefix)
		}
	}
}

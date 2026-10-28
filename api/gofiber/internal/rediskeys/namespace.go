package rediskeys

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var validNamespace = regexp.MustCompile("^[a-z0-9][a-z0-9_-]*$")

// Namespace is read after dotenv loading, never during package initialization.
// The fallback preserves Classic's existing keys for local/legacy callers.
func Namespace() string {
	value := os.Getenv("REDIS_NAMESPACE")
	if value == "" {
		return "grc"
	}
	return value
}

func Validate(value string) error {
	if !validNamespace.MatchString(value) {
		return fmt.Errorf("REDIS_NAMESPACE must contain only lowercase letters, digits, underscores or hyphens and start with a letter or digit")
	}
	return nil
}

func Key(logical string) string      { return Namespace() + "/" + logical }
func Logical(physical string) string { return strings.TrimPrefix(physical, Namespace()+"/") }

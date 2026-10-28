package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"github.com/gofiber/fiber/v2"
	"os"
	"strings"
)

const adminAPITokenLocalKey = "verified_admin_api_token"

func acceptsAdminAPIToken(route, token string) bool {
	configured := os.Getenv("ADMIN_API_TOKEN")
	if configured == "" || token == "" {
		return false
	}
	adminRoute := false
	for _, segment := range strings.Split(route, "/") {
		if segment == "admin" {
			adminRoute = true
			break
		}
	}
	if !adminRoute {
		return false
	}
	expected, actual := sha256.Sum256([]byte(configured)), sha256.Sum256([]byte(token))
	return subtle.ConstantTimeCompare(expected[:], actual[:]) == 1
}

// Group middleware sees the parent group's route, not the final endpoint.
// Match against registered static admin prefixes rather than arbitrary URL parameters.
func isRegisteredAdminRequest(c *fiber.Ctx) bool {
	for _, route := range c.App().GetRoutes() {
		parts := strings.Split(route.Path, "/")
		for i, part := range parts {
			if strings.ContainsAny(part, ":*+") {
				break
			}
			if part != "admin" {
				continue
			}
			prefix := strings.Join(parts[:i+1], "/")
			if c.Path() == prefix || strings.HasPrefix(c.Path(), prefix+"/") {
				return true
			}
			break
		}
	}
	return false
}

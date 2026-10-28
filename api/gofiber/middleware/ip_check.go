package middleware

import (
	"context"
	myMDL "gofiber/model"
	myREPO "gofiber/repository"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func IPCheckMiddleware(ipcRepo myREPO.IPCheckRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientIP := c.IP()

		// Guard against empty IP string if proxy validation fails
		if clientIP == "" {
			// Fallback: Use remote connection IP, or fail/log depending on security requirements
			clientIP = c.Context().RemoteIP().String()
		}

		// Handle comma-separated IPs if Fiber returns a raw proxy list string
		if strings.Contains(clientIP, ",") {
			clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
		}

		// // Diagnostic log: Shows both raw Fiber resolution and final sanitized target IP
		// log.Printf("[IP Check] Target IP: %q | Fiber IP(): %q | IPs(): %v | Remote Socket: %s",
		// 	clientIP,
		// 	c.IP(),
		// 	c.IPs(),
		// 	c.Context().RemoteIP().String(),
		// )

		// Optional Guard: Prevent querying DB for local loopback / unroutable IPs
		if clientIP == "127.0.0.1" || clientIP == "::1" || clientIP == "" {
			return c.Next()
		}

		// Add a quick timeout so slow DB queries don't hang the request chain
		// Use c.UserContext() or context.Background() for database timeouts in Fiber
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()

		isBanned, err := ipcRepo.IsBanned(ctx, clientIP)
		if err != nil {
			log.Printf("[IP Check] DB error: %v, passing request to avoid false positive.", err)
			// Log error, but proceed (or fail closed based on your security policy)
			return c.Next()
		}

		if isBanned {
			// log.Printf("[IP Check] Blocked IP: %s", clientIP)
			//TODO: report? log this ip?
			resp := myMDL.Default403Resp.Copy()
			resp.SetErrorAndMessage("Access denied: Your IP is unavailable.")
			return c.Status(resp.Status).JSON(resp)
		}

		return c.Next()
	}
}

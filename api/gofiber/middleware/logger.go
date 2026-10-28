package middleware

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var ConfigDefault = logger.Config{
	Next:         nil,
	Format:       "[${time}] ${status} - ${latency} ${method} ${url}\n",
	TimeFormat:   "15:04:05",
	TimeZone:     "Local",
	TimeInterval: 500 * time.Millisecond,
	Output:       os.Stderr,
}

func BuildLogger() fiber.Handler {
	return logger.New(ConfigDefault)
}

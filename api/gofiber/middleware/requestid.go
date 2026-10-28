package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/utils"
)

var (
	requestIDConfigDefault = requestid.Config{
		Next:   nil,
		Header: fiber.HeaderXRequestID,
		Generator: func() string {
			return utils.UUID()
		},
		ContextKey: "reqid",
	}
)

func BuildRequestID() fiber.Handler {
	return requestid.New(requestIDConfigDefault)
}

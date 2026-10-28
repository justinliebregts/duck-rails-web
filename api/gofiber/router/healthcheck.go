package router

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	myHDLR "gofiber/handler"
	"log"
)

func SetupHealthCheckRoutes(ap *fiber.App) {

	// Health
	if hl, err := myHDLR.HealthCheckHandler(); err == nil {
		log.Println("SetupHealthCheckRoutes: register GET /live and /ready")
		ap.Get("/live", adaptor.HTTPHandlerFunc(hl.LiveEndpoint))
		ap.Get("/ready", adaptor.HTTPHandlerFunc(hl.ReadyEndpoint))

		if myHDLR.EnablePrometheus {
			ap.Get("/metrics", adaptor.HTTPHandler(myHDLR.PromhttpHandlerFor))
		}
	} else {
		log.Println("SetupHealthCheckRoutes: HealthCheck Skipped!")
	}
}

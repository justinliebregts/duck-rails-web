package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/expvar"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"gofiber/config"
	"log"
)

func SetupDevInfo(ap *fiber.App) {
	if config.IsEnvDevelopment() {
		log.Println("SetupAppMonitorRoutes: register GET /debug/vars and /debug/pprof/")
		ap.Use(expvar.New()) // /debug/vars
		ap.Use(pprof.New())  // /debug/pprof/
	}
}

func SetupAppMonitorRoutes(ap *fiber.App) {
	if config.IsMonitorEnabled() {
		log.Println("SetupAppMonitorRoutes: register GET /monitor")
		ap.Get("/monitor", monitor.New(monitor.Config{Title: "Gun Raiders Classic API Monitor Page"}))
	}
}

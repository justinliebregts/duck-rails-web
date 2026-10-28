package router

import (
	"github.com/gofiber/fiber/v2"
	myHDR "gofiber/handler"
)

func SetupPhotonRoutes(rtr fiber.Router, session myHDR.Session) {
	rtr.Get("/api/photon/authenticate", session.SessionCheckHandler)
	rtr.Get("/api/fusion/authenticate", session.SessionCheckHandler)

	rtr.Get("/api/v99/photon/authenticate", session.SessionCheckHandler)
	rtr.Get("/api/v99/fusion/authenticate", session.SessionCheckHandler)
}

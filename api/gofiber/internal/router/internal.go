package router

import (
	itHDL "gofiber/internal/handler"
	myMW "gofiber/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupInternalRoutes(rtr fiber.Router) {

	// internal handlers
	internalHDL := itHDL.NewInternalHandler()

	// everything under this endpoint only accessible by local ip
	internalRtr := rtr.Group("/internal")
	internalRtr.Get("/api-version", internalHDL.GetAPIVersion)

	itlPxyRtr := internalRtr.Group("/proxy")
	itlPxyRtr.Get("/status", internalHDL.ProxyInfoGet)

	itlPxyAdminRtr := itlPxyRtr.Group("/admin", myMW.CustomOAuth2Protected, myMW.AdminPrivilegeCheckMiddleware)
	itlPxyAdminRtr.Get("/*", internalHDL.ProxyRevert)
}

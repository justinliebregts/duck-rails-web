package router

import (
	"github.com/gofiber/fiber/v2"
	myHDR "gofiber/handler"
	"gofiber/middleware"
)

func SetupBuildHashRoutes(rtr fiber.Router) {
	bhRtr := rtr.Group("/buildhash", middleware.CustomOAuth2Protected)
	bhRtr.Post("/validate", myHDR.BuildHashValidateHandler)

	bhRtrAdm := bhRtr.Group("/admin", middleware.AdminPrivilegeCheckMiddleware)
	bhRtrAdm.Post("/create", myHDR.BuildHashAdminCreateHandler)
}

package router

import (
	"github.com/gofiber/fiber/v2"
	myHDR "gofiber/handler"
	"gofiber/middleware"
)

func SetupSkinRoutes(rtr fiber.Router) {
	apiRouter := rtr.Group("/skin", middleware.CustomOAuth2Protected)
	// no player function here

	apiRtrAdm := apiRouter.Group("/admin", middleware.AdminPrivilegeCheckMiddleware)
	apiRtrAdm.Get("/all", func(c *fiber.Ctx) error { return c.RedirectToRoute("skin-all", fiber.Map{}) })
	apiRtrAdm.Get("/", myHDR.SkinAdminGetAll).Name("skin-all")
	apiRtrAdm.Post("/create", myHDR.SkinAdminCreate)
	apiRtrAdm.Post("/update/:skin_id", myHDR.SkinAdminUpdate)
}

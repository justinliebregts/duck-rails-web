package router

import (
	myHDR "gofiber/handler"
	"gofiber/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupItemRoutes(rtr fiber.Router) {
	apiRtr := rtr.Group("/item", middleware.CustomOAuth2Protected)
	apiRtrAdm := apiRtr.Group("/admin", middleware.AdminPrivilegeCheckMiddleware)
	apiRtrAdm.Get("/all", myHDR.ItemGetAllHandler)
	apiRtrAdm.Get("/weapons", myHDR.ItemListWeaponsHandler)
	apiRtrAdm.Post("/create", myHDR.ItemPostCreateHandler)
	apiRtrAdm.Post("/update/:item_id<int>", myHDR.ItemPostUpdateHandler)
}

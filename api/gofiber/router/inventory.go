package router

import (
	myHDR "gofiber/handler"
	myMW "gofiber/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupInventoryRoutes(rtr fiber.Router) {
	apiRtr := rtr.Group("/inventory", myMW.CustomOAuth2Protected)
	apiRtr.Get("/all", myHDR.AllInventoryHandler)
	apiRtr.Get("/all/:dev_owned<int>", myHDR.AllInventoryHandler)
	apiRtr.Post("/buy", myHDR.BuyInventoryHandler)

	apiRtrAdm := apiRtr.Group("/admin", myMW.AdminPrivilegeCheckMiddleware)
	apiRtrAdm.Get("/all/:player_id", myHDR.AllAdminInventoryHandler)
	apiRtrAdm.Post("/give-item", myHDR.GiveItemInventoryHandler)
	apiRtrAdm.Post("/delete/:transaction_id<int>", myHDR.DeleteAdminInventoryHandler)
	apiRtrAdm.Post("/delete-all", myHDR.DeleteAllAdminInventoryHandler)
}

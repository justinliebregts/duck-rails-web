package router

import (
	myHDR "gofiber/handler"
	"gofiber/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupPlayerRoutes(rtr fiber.Router) {
	apiRouter := rtr.Group("/player")

	apiRouter.Post("/heartbeat", middleware.CustomOAuth2Protected, myHDR.PlayerPostHeartbeat)

	apiRtrAdm := apiRouter.Group("/admin", middleware.CustomOAuth2Protected, middleware.AdminPrivilegeCheckMiddleware)
	apiRtrAdm.Get("/online", myHDR.PlayerListAlive)
	apiRtrAdm.Post("/cease/:player_uuid", myHDR.PlayerPostTerminateHeartbeat)

	apiRtrAdm.Get("/all", myHDR.PlayerGetAllAdminHandler)
	apiRtrAdm.Get("/get/:player_id", func(c *fiber.Ctx) error {
		return c.RedirectToRoute("player-admin-get", fiber.Map{"queries": map[string]string{"player_id": c.Params("player_id")}})
	})
	apiRtrAdm.Get("/get", myHDR.PlayerGetAdminHandler).Name("player-admin-get")
	apiRtrAdm.Get("/search", myHDR.PlayerSearchAdminHandler).Name("player-admin-search")
	apiRtrAdm.Get("/search-by-id/:player_id", func(c *fiber.Ctx) error {
		return c.RedirectToRoute("player-admin-search", fiber.Map{"queries": map[string]string{"player_id": c.Params("player_id")}})
	})
	apiRtrAdm.Get("/search-by-name/:player_name", func(c *fiber.Ctx) error {
		return c.RedirectToRoute("player-admin-search", fiber.Map{"queries": map[string]string{"player_name": c.Params("player_name")}})
	})

}

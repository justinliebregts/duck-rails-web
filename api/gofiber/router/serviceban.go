package router

import (
	myHDR "gofiber/handler"
	"gofiber/middleware"
	"gofiber/repository"
	"gofiber/service/cloudflare"

	"github.com/gofiber/fiber/v2"
)

func SetupServiceBanRoutes(rtr fiber.Router, ipcRepo repository.IPCheckRepository, cloudflareService cloudflare.CloudflareIPBanService) {

	apiRtr := rtr.Group("/serviceban", middleware.CustomOAuth2Protected)
	apiRtr.Get("/get", myHDR.ServiceBanListActiveByPlayerHandler).Name("ban-get-v2") // Authenticated player only

	apiRtrAdm := apiRtr.Group("/admin", middleware.AdminPrivilegeCheckMiddleware)
	apiRtrAdm.Get("/all", myHDR.ServiceBanListAllAdminHandler).Name("service-ban-admin-all")
	apiRtrAdm.Get("/get/:gr_uuid", myHDR.ServiceBanGetPlayerAdminHandler)
	apiRtrAdm.Post("/create", myHDR.ServiceBanPlayerAdminHandler)
	apiRtrAdm.Post("/end/:serban_id", myHDR.EndServiceBanAdminHandler)

	apiRtrAdm.Post("/ip-blacklist", myHDR.ServiceBanIPBlacklistAdminCreateHandler(ipcRepo))
	apiRtrAdm.Get("/ip-blacklist", myHDR.ServiceBanIPBlacklistAdminHandler(ipcRepo))
	apiRtrAdm.Put("/ip-blacklist", myHDR.ServiceBanIPBlacklistAdminRecoverBanHandler(ipcRepo))
	apiRtrAdm.Delete("/ip-blacklist", myHDR.ServiceBanIPBlacklistAdminDeleteHandler(ipcRepo))
	//apiRtrAdm.Put("/ip-blacklist/:id", myHDR.ServiceBanIPBlacklistAdminHandler(ipcRepo))

	apiRtrAdm.Get("/cloudflare-ip-blacklist", myHDR.ServiceBanCloudflareIPBlacklistAdminHandler(cloudflareService))
}

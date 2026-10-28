package router

import (
	myHDR "gofiber/handler"
	myMW "gofiber/middleware"
	myREPO "gofiber/repository"

	"github.com/gofiber/fiber/v2"
)

func SetupSessionRoutes(rtr fiber.Router, sesHdl myHDR.Session, ipCheckRepo myREPO.IPCheckRepository) {

	apiRouter := rtr.Group("/session")

	apiRouter.Post("/create", myMW.IPCheckMiddleware(ipCheckRepo), sesHdl.SessionCreateHandler)
	//apiRouter.Post("/refresh", middleware.CustomOAuth2Protected, myHDR.SessionRefreshHandler) // not used ATM

	apiRouter.Post("/:session_id", myMW.IPCheckMiddleware(ipCheckRepo), sesHdl.SessionRefreshHandler)
	apiRouter.Get("/report", sesHdl.SessionPlatformStatsHandler)
	apiRouter.Get("/online", sesHdl.SessionActiveListHandler)

	admRtr := apiRouter.Group("/admin", myMW.CustomOAuth2Protected, myMW.AdminPrivilegeCheckMiddleware)
	admRtr.Get("/:platform/:platform_id", sesHdl.AdminGetAllTokens) // list all access/ refresh tokens related to this user
	admRtr.Delete("/:platform/:platform_id", sesHdl.AdminRemoveAllTokens)
}

//func SetupPhotonSessionCheckRoutes(rtr fiber.Router) {
//	rtr.Get("/photon/authenticate", myHDR.SessionCheckHandler)
//	rtr.Get("/fusion/authenticate", myHDR.SessionCheckHandler)
//}

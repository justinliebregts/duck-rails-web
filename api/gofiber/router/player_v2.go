package router

import (
	v2 "gofiber/handler/v2"
	"gofiber/internal/service/remote/oculus"
	myMW "gofiber/middleware"
	myREPO "gofiber/repository"
	mySRV "gofiber/service"

	"github.com/gofiber/fiber/v2"
)

func SetupPlayerV2Routes(rtr fiber.Router, playerRepo myREPO.PlayerRepository, ipCheckRepo myREPO.IPCheckRepository, playerSvc mySRV.Player, oculusRemoteSRV oculus.Client) {
	playerV2HDL := v2.NewPlayerV2Handler(playerRepo, playerSvc, oculusRemoteSRV)

	// Public (no auth required); Meta Quest Attestation protection before backend registration
	publicPlayerRoutes := rtr.Group("/player/v2")
	publicPlayerRoutes.Post("/social-account/oculus", myMW.IPCheckMiddleware(ipCheckRepo), myMW.MetaQuestAttestationMiddleware(oculusRemoteSRV), playerV2HDL.PlayerSocialAccountHandler)

	// player v2 admin
	admPlayerRoutes := publicPlayerRoutes.Group("/admin", myMW.CustomOAuth2Protected, myMW.AdminPrivilegeCheckMiddleware)
	admPlayerRoutes.Get("/oculus-org-scoped-id", playerV2HDL.GetOculusOrgScopedIDHandler)
	admPlayerRoutes.Get("/list", playerV2HDL.PlayerAdminListHandler)

	// player v2 admin oculus attestation
	admPlayerOculusRoutes := admPlayerRoutes.Group("/oculus-attestation")
	admPlayerOculusRoutes.Get("/list", playerV2HDL.ListAttestationRecordHandler)
	admPlayerOculusRoutes.Post("/toggle-ban", playerV2HDL.PlayerOculusAttestationToggleBanHandler)
	admPlayerOculusRoutes.Get("/ban-list", playerV2HDL.ListAttestationDeviceBanRecordHandler)
	admPlayerOculusRoutes.Get("/remote-ban-status/:unique_id", playerV2HDL.CheckRemoteAttestationBanStatus)
}

package duckz

import (
	myCFG "gofiber/config"
	myDB "gofiber/database/mysql"
	dzHDL "gofiber/handler/duckz"
	oculus "gofiber/internal/service/remote/oculus"
	myMW "gofiber/middleware"
	dzREPO "gofiber/repository/duckz"
	dzSRV "gofiber/service/duckz"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// SetupDuckzEconomyRoutes registers the reused economy under the Rails API root.
func SetupDuckzEconomyRoutes(rtr fiber.Router, oculusSvc oculus.Client) {
	eco, err := myCFG.LoadDuckzEconomyJson()
	if err != nil {
		log.Fatalf("duckz economy config: %v", err)
	}

	repo := dzREPO.NewEconomyRepository(myDB.DBConn)
	verifier := dzSRV.NewMetaEntitlementVerifier(oculusSvc)
	svc := dzSRV.NewEconomyService(repo, eco, verifier)
	hdl := dzHDL.NewEconomyHandler(svc)

	// Bind authentication to each endpoint: an empty-prefix group would also
	// protect unrelated public routes registered later under /api/v1.
	rtr.Get("/wallet", myMW.CustomOAuth2Protected, hdl.Wallet)
	rtr.Get("/store", myMW.CustomOAuth2Protected, hdl.Store)
	rtr.Post("/coins/deposit", myMW.CustomOAuth2Protected, myMW.NonceRepeatCheck(), hdl.Deposit)
	rtr.Post("/store/buy", myMW.CustomOAuth2Protected, myMW.NonceRepeatCheck(), hdl.Buy)
	rtr.Post("/iap/grant", myMW.CustomOAuth2Protected, myMW.NonceRepeatCheck(), hdl.GrantIAP)
	rtr.Post("/loadout", myMW.CustomOAuth2Protected, myMW.NonceRepeatCheck(), hdl.SaveLoadout)

	// Admin-only. Privilege is checked by AdminPrivilegeCheckMiddleware.
	dzAdm := rtr.Group("/admin", myMW.CustomOAuth2Protected, myMW.AdminPrivilegeCheckMiddleware)
	dzAdm.Get("/wallet/:player_id", hdl.AdminWallet)
	dzAdm.Post("/coins/adjust", hdl.AdminAdjust)
	dzAdm.Get("/transactions/:player_id", hdl.AdminTransactions)
	dzAdm.Get("/check", hdl.AdminCheck)
}

package router

import (
	"github.com/gofiber/fiber/v2"
	myHDR "gofiber/handler"
	myMW "gofiber/middleware"
	myREPO "gofiber/repository"
)

func SetupPatchRoutes(rtr fiber.Router, dbPatchRepo myREPO.DBPatchRepository) {
	patchHDL := myHDR.NewPatchHandler(dbPatchRepo)

	apiRouter := rtr.Group("/patch")

	apiRtrAdm := apiRouter.Group("/admin", myMW.CustomOAuth2Protected, myMW.AdminPrivilegeCheckMiddleware)
	apiRtrAdm.Post("rename", patchHDL.PatchAccountRenameHandler)
}

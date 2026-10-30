package router

import (
	myCFG "gofiber/config"
	myHDR "gofiber/handler"
	"gofiber/internal/router"
	oculusRmtSRV "gofiber/internal/service/remote/oculus"
	myREPO "gofiber/repository"
	mySRV "gofiber/service"
	"gofiber/service/cloudflare"
	myK8sSRV "gofiber/service/k8s"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"k8s.io/client-go/kubernetes"
)

func RegisterInternalRoutes(ap *fiber.App) {
	router.SetupInternalRoutes(ap)
}

func RegisterRoutesByApp(ap *fiber.App, sesHdl myHDR.Session) {

	// Register HealthCheck
	SetupHealthCheckRoutes(ap)

	// Register Monitor
	SetupAppMonitorRoutes(ap)

	// Register Development Info
	SetupDevInfo(ap)

	// special case for /api/photon/authenticate and /api/fusion/authenticate
	if myCFG.IsPhotonSessionCheckEnabled() {
		log.Info("PhotonSessionCheck Enabled")
		SetupPhotonRoutes(ap, sesHdl)
	}
}

type AppContainer struct {
	// Repositories
	DbPatchRepo myREPO.DBPatchRepository
	IPCheckRepo myREPO.IPCheckRepository
	PlayerRepo  myREPO.PlayerRepository
	SessionRepo myREPO.Session

	// Services
	CronJobSvc    myK8sSRV.CronJobService
	PlayerSvc     mySRV.Player
	SessionSvc    mySRV.Session
	CloudflareSvc cloudflare.CloudflareIPBanService

	OculusRemoteSvc oculusRmtSRV.Client

	K8sClient kubernetes.Interface
}

func RegisterAPIRoutes(apir fiber.Router, ac *AppContainer, sesHdl myHDR.Session) {
	SetupDiagnosticRoutes(apir)
	SetupBuildHashRoutes(apir)
	SetupInfoRoutes(apir, ac.CronJobSvc)
	SetupPatchRoutes(apir, ac.DbPatchRepo)
	SetupPlayerRoutes(apir)
	SetupPlayerV2Routes(apir, ac.PlayerRepo, ac.IPCheckRepo, ac.PlayerSvc, ac.OculusRemoteSvc)
	SetupServiceBanRoutes(apir, ac.IPCheckRepo, ac.CloudflareSvc)
	SetupSessionRoutes(apir, sesHdl, ac.IPCheckRepo)

	SetupInventoryRoutes(apir)
	SetupItemRoutes(apir)
	SetupSkinRoutes(apir)
}

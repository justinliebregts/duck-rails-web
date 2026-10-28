package router

import (
	myHDR "gofiber/handler"
	myMW "gofiber/middleware"
	myK8sSRV "gofiber/service/k8s"

	"github.com/gofiber/fiber/v2"
)

func SetupInfoRoutes(rtr fiber.Router, myK8sSvc myK8sSRV.CronJobService) {
	apiRtr := rtr.Group("/info")

	apiRtr.Get("/", func(c *fiber.Ctx) error { return c.RedirectToRoute("info-all", fiber.Map{}) })
	apiRtr.Get("/all", myMW.PageCachedWithCacheDuration(600), myHDR.InfoAllHandler).Name("info-all")
	apiRtr.Get("/version-:device_platform", func(c *fiber.Ctx) error {
		return c.RedirectToRoute("info-version", fiber.Map{"queries": map[string]string{"device_platform": c.Params("device_platform")}})
	})
	apiRtr.Get("/version", myHDR.InfoVersionHandler).Name("info-version")
	apiRtr.Get("/time", myHDR.InfoTimeHandler).Name("info-time")

	apiRtrAdm := apiRtr.Group("/admin", myMW.CustomOAuth2Protected, myMW.AdminPrivilegeCheckMiddleware)
	apiRtrAdm.Get("/all", myHDR.InfoAllHandler)
	apiRtrAdm.Get("/env", myHDR.InfoOSEnvironmentHandler)

	//apiRtrAdm.Post("/hashcheck", myHDR.InfoSetHashCheckHandler)
	apiRtrAdm.Get("/hashcheck/list", myHDR.InfoGetHashCheckIgnoredHandler)
	apiRtrAdm.Get("/hashcheck", myHDR.InfoGetHashCheckHandler)

	//k8s related
	k8sHDL := myHDR.NewK8sHandler(myK8sSvc)
	apiRtrAdm.Get("/namespaces/:ns/jobs/:jb/logs", k8sHDL.RetrieveJobLogHandler).Name("info-k8s-jobs-log")
	apiRtrAdm.Get("/namespaces/:ns/jobs", k8sHDL.ListExecutedJobsHandler).Name("info-k8s-jobs")
	apiRtrAdm.Post("/namespaces/:ns/cronjobs/:cj/trigger", k8sHDL.TriggerCronJobHandler).Name("info-k8s-cronjobs-trigger")
	apiRtrAdm.Get("/namespaces/:ns/cronjobs", k8sHDL.ListCronJobsHandler).Name("info-k8s-cronjobs")
}

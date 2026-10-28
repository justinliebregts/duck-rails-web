package main

import (
	"fmt"
	mySTO "gofiber/cache"
	myCFG "gofiber/config"
	authCFG "gofiber/config/remote_auth"
	"gofiber/database/mysql"
	myHDL "gofiber/handler"
	"gofiber/internal/rediskeys"
	oculusRmtSRV "gofiber/internal/service/remote/oculus"
	myMW "gofiber/middleware"
	myDOM "gofiber/model"
	cfgMDL "gofiber/model/config"
	myREPO "gofiber/repository"
	myRTR "gofiber/router"
	mySRV "gofiber/service"
	"gofiber/service/cloudflare"
	myK8sSRV "gofiber/service/k8s"
	"log"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func main() {

	//Config from .env then os
	myCFG.LoadDotEnv(filepath.Base("."))
	if err := myCFG.ValidateProductionConfiguration(); err != nil {
		log.Fatal(err)
	}

	//myCFG.LoadParametersFromOS()

	if err := rediskeys.Validate(rediskeys.Namespace()); err != nil {
		log.Fatal(err)
	}

	// Cache/Redis
	setupCache()

	//load OCULUS CFG
	ocuCFG := authCFG.LoadOculusConfig()
	ocuRemoteSrv := oculusRmtSRV.NewOculusService(ocuCFG)

	// Database/MySQL
	dbCFG := myCFG.LoadDBConfig()
	dbRW := setupDatabase(dbCFG)

	sesRepo := myREPO.NewSession()
	dbPatchRepo := myREPO.NewDBPatchRepository(dbRW)
	playerRepo := myREPO.NewPlayerRepository(dbRW)

	playerSRV := mySRV.NewPlayerService(playerRepo, dbPatchRepo)

	sesSvc := mySRV.NewSessionService(sesRepo)

	// if expose db through prometheus
	myCFG.SetupDBProm()

	// load config from env; set into model through handler
	myHDL.SetPlayerConfig(myCFG.RetrieveHeartBeatFrequencyAndLifeCycle())

	// setup hash provider used in password hash
	myCFG.SetupHashProvider()

	// Setup CustomError Handler
	myCFG.SetCustomErrorHandler(myHDL.LoadCustomErrorHandler)

	// Create Dev Session in Redis
	if myCFG.IsEnvDevelopment() {
		myHDL.CreateDevSession()
		myHDL.GiveDevSessionAdminPrivilege()
	}

	myDOM.LoadSessionSettingsFromEnv()

	app := fiber.New(myCFG.FiberAppConfig)

	// Middleware to set a value in the context
	app.Use(func(c *fiber.Ctx) error {
		baseConfiguration := myCFG.BaseConfiguration{}
		baseConfiguration.Load()

		// Set a value in the context
		c.Locals("base_configuration", baseConfiguration)
		return c.Next()
	})

	app.Use(recover.New())

	// allow Access-Control-Allow-Origin in dev build
	if myCFG.IsEnvDevelopment() {

		corOrigins := "*"
		log.Printf("Allowing COR for %v", corOrigins)

		app.Use(cors.New(cors.Config{
			AllowOrigins: corOrigins,
			AllowMethods: "GET,POST,PUT,PATCH,DELETE",
			AllowHeaders: "Content-Type,Authorization,Access-Control-Allow-Origin,If-None-Match",
			//ExposeHeaders:    "X-Custom-Header",
			AllowCredentials: false,
		}))
	}

	sesHdl := myHDL.NewSessionHandler(sesRepo, sesSvc)
	// Register Routes by
	myRTR.RegisterRoutesByApp(app, sesHdl)

	// custom logger
	app.Use(myMW.BuildLogger())

	// Setup RequestID
	setupRequestID(app)

	//Enable eTag
	myCFG.EnableETag(app)
	myCFG.PrintCachedPageList()

	//Enable Page Cache for Certain HEAD and GET; POST, UPDATE, DELETE will never be cached!
	myCFG.ActivatePageCache(app, mySTO.DBConn)

	// root
	app.Head("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	//app.Get("/", func(c *fiber.Ctx) error {
	//	return c.Status(fiber.StatusOK).SendString("OK")
	//})

	// internal private ip only endpoints
	myRTR.RegisterInternalRoutes(app)

	apiRouter := app.Group(filepath.Join("/api", myCFG.RetrieveApiVersion()))

	k8sClient, err := initK8sClient()
	if err != nil {
		log.Fatalf("failed to initialize k8s client: %v", err)
	}

	resourceContainer := myRTR.AppContainer{
		DbPatchRepo: dbPatchRepo,
		IPCheckRepo: myREPO.NewIPCheckRepository(dbRW),
		PlayerRepo:  playerRepo,
		SessionRepo: sesRepo,

		CronJobSvc: myK8sSRV.NewCronJobService(k8sClient),
		PlayerSvc:  playerSRV,
		SessionSvc: sesSvc,

		CloudflareSvc: cloudflare.NewCloudflareIPBanService(
			myCFG.CloudflareAPIToken(),
			myCFG.CloudflareZoneID(),
		),

		OculusRemoteSvc: ocuRemoteSrv,

		K8sClient: k8sClient,
	}

	// Register Routes
	myRTR.RegisterAPIRoutes(apiRouter, &resourceContainer, sesHdl)

	// Provide a minimal config
	app.Use(favicon.New())

	// JWT Middleware
	setupJWTMiddleware(app)

	// any route registered after JWT will be protected

	log.Fatal(app.Listen(":3000"))
}

//// loadTemplateEngine Load Template for HTML display; Not use at this moment
//func loadTemplateEngine() *html.Engine {
//	// Initialize standard Go html template engine
//	return html.New("./views", ".html")
//}

func setupRequestID(ap *fiber.App) {
	if myCFG.IsRequestIDEnabled() {
		log.Println("SetupRequestID: Enable RequestID in Response Header")
		ap.Use(myMW.BuildRequestID())
	}
}

func setupDatabase(dbCfg cfgMDL.DBConfig) *gorm.DB {
	mysql.SetupGORMLogLevel(myCFG.IsEnvDevelopment())
	return mysql.NewDatabase(dbCfg)
}

func setupCache() {
	cnf := myCFG.LoadCacheConfig()

	if myCFG.IsEnvDevelopment() {
		log.Printf("Cache config from ENV:%v\n", cnf)
		log.Printf("default Redis config:%v\n", mySTO.DebugPrintConfig())
	}

	mySTO.UpdateConfig(cnf)

	if myCFG.IsEnvDevelopment() {
		log.Printf("Redis using config:%v\n", mySTO.DebugPrintConfig())
	}

	mySTO.ConnectDb()
}

func setupJWTMiddleware(app *fiber.App) {
	_, err := myCFG.SetupRSAPrivateKey()
	if err != nil {
		log.Println("JWT is Disabled!")
		return
	}

	app.Use(myCFG.NewJWTHandler())
}

func initK8sClient() (kubernetes.Interface, error) {

	if myCFG.IsFakeLocalK8S() {
		log.Println("Initializing fake in-memory K8s client for local dev/testing")
		return newTestK8sClient(), nil
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("in-cluster config error: %w", err)
	}
	return kubernetes.NewForConfig(config)
}

func newTestK8sClient() kubernetes.Interface {
	return fake.NewSimpleClientset()
}

package handler

import (
	"gofiber/config"
	"gofiber/model"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type InfoAllResponse struct {
	Version            string                  `json:"version"`
	JwtInfo            *map[string]interface{} `json:"jwt,omitempty"`
	HeartbeatFrequency float64                 `json:"hartbet_freq,omitempty"`
}

// InfoAllHandler godoc
// @Summary Get the app info
// @Description
// @Tags info
// @Accept */*
// @Produce json
// @Success 200 {object} model.ResponseGeneral{status=int,data=InfoAllResponse}
// @Router /info [get]
// @Router /info/all [get]
// @Router /info/admin/all [get]
func InfoAllHandler(c *fiber.Ctx) (err error) {
	ji := map[string]interface{}(config.FormJWTInfo())

	resp := model.Default200Resp.Copy()
	resp.Data = InfoAllResponse{
		Version:            config.RetrieveAppVersion("OCULUS"),
		JwtInfo:            &ji,
		HeartbeatFrequency: model.HeartbeatFrequency.Seconds(),
	}
	return c.Status(resp.Status).JSON(resp)
}

// InfoVersionHandler godoc
// @Summary Get the version number
// @Description
// @Tags info
// @Accept */*
// @Param	device_platform	query	string	false	"platform"	Enums(OCULUS, IQIYI, PICO, YVR)
// @Produce text/plain
// @Success 200 string string
// @Router /info/version [get]
// @Router /info/version-pico [get]
// @Router /info/version-iqiyi [get]
// @Router /info/version-google [get]
// @Router /info/version-android [get]
// @Router /info/version-apple [get]
// @Router /info/version-ios [get]
func InfoVersionHandler(c *fiber.Ctx) (err error) {
	q := strings.ToUpper(c.Query("device_platform", "OCULUS"))
	return c.Status(fiber.StatusOK).SendString(config.RetrieveAppVersion(q))
}

// InfoOSEnvironmentHandler godoc
// @Summary Expose OS Environment
// @Description
// @Tags info
// @Accept */*
// @Produce text/plain
// @Success 200 str string
// @Router /info/admin/env [get]
func InfoOSEnvironmentHandler(c *fiber.Ctx) (err error) {
	return c.Status(fiber.StatusOK).
		SendString(strings.Join(config.RetrieveOSEnvironment(), "\n"))
}

// InfoTimeResponse is UTC server time for clients (edge-cached ~1s).
type InfoTimeResponse struct {
	Unix int64  `json:"unix"`
	ISO  string `json:"iso"`
}

// InfoTimeHandler godoc
// @Summary Get server UTC time
// @Description Returns current UTC time. Origin sets Cache-Control: public, max-age=1 for Cloudflare edge cache (requires a Cache Rule; no Redis).
// @Tags info
// @Accept */*
// @Produce json
// @Success 200 {object} model.ResponseGeneral{status=int,data=InfoTimeResponse}
// @Router /info/time [get]
func InfoTimeHandler(c *fiber.Ctx) error {
	now := model.CurrentTimestamp()
	setPublicCache(c, 1)
	resp := model.Default200Resp.Copy()
	resp.Data = InfoTimeResponse{
		Unix: now.Unix(),
		ISO:  now.Format(time.RFC3339),
	}
	return c.JSON(resp)
}

// InfoGetHashCheckIgnoredHandler godoc
// @Summary Get the MismatchHashCheck ignore settings
// @Description
// @Tags info
// @Accept */*
// @Produce application/json
// @Success 200 {object} []string
// @Router /info/admin/hashcheck/list [get]
func InfoGetHashCheckIgnoredHandler(c *fiber.Ctx) (err error) {
	//ignored := config.RetrieveCurrentBuildHashCheckIgnored()
	//if ignored == nil {
	//	ignored = []string{}
	//}
	//return c.JSON(fiber.Map{"ignored_platforms": ignored})

	toCheck := config.RetrieveCurrentBuildHashCheck()
	if toCheck == nil {
		toCheck = []string{}
	}
	return c.JSON(fiber.Map{"buildhash_to_check_platforms": toCheck})
}

// InfoGetHashCheckHandler godoc
// @Summary Get if BuildHash Check
// @Description
// @Tags info
// @Accept */*
// @Param pla query string true "Platform Name"
// @Produce application/json
// @Success 200 {object} string
// @Router /info/admin/hashcheck [get]
func InfoGetHashCheckHandler(c *fiber.Ctx) (err error) {
	//platform := strings.ToUpper(c.Params("platform")) // Get the platform from the URL path
	platform := strings.ToUpper(c.Query("pla"))

	if platform == "" {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]string{"error": "Platform is required"})
	}

	return c.JSON(fiber.Map{
		"platform":           platform,
		"buildhash_to_check": config.IsPlatformBuildHashToCheck(platform),
	})
}

func setPublicCache(c *fiber.Ctx, maxAge int) {
	if maxAge < 1 {
		maxAge = 1
	}
	c.Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
}

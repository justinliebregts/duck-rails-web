package internal

import (
	myCFG "gofiber/config"
	myMDL "gofiber/model"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

type Internal interface {
	GetAPIVersion(c *fiber.Ctx) error

	ProxyInfoGet(c *fiber.Ctx) error
	ProxyRevert(c *fiber.Ctx) error
}

func NewInternalHandler() Internal {
	return &internalHandlerImpl{}
}

type internalHandlerImpl struct {
}

// GetAPIVersion godoc
// @Summary get api version
// @Description get api version
// @Tags internal
// @Produce text/plain
// @Success 200 {string} string "api version"
// @Failure 500 {string} string "api version"
// @Router /internal/api-version [get]
func (r *internalHandlerImpl) GetAPIVersion(c *fiber.Ctx) error {
	// return c.JSON(fiber.Map{
	// 	"api_version": myCFG.RetrieveApiVersion,
	// })
	return c.SendString(myCFG.RetrieveApiVersion())
}

// ProxyInfoGet godoc
// @Summary get proxy status info
// @Description get proxy status info
// @Tags internal
// @Produce application/json
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /internal/proxy/status [get]
func (r *internalHandlerImpl) ProxyInfoGet(c *fiber.Ctx) error {
	resp := myMDL.Default200Resp.Copy()

	configK8sServiceURL := myCFG.RetrieveK8sServiceURL()

	// Quick connectivity check
	targetOnline := false
	targetReady := false
	client := &http.Client{Timeout: 800 * time.Millisecond}
	if diagResp, err := client.Get(configK8sServiceURL + "/live"); err == nil {
		targetOnline = diagResp.StatusCode == http.StatusOK
		diagResp.Body.Close()
	}

	if diagResp2, err := client.Get(configK8sServiceURL + "/ready"); err == nil {
		targetReady = diagResp2.StatusCode == http.StatusOK
		diagResp2.Body.Close()
	}

	targetAPIVersion := "UnKnOwN"
	if diagResp3, err := client.Get(configK8sServiceURL + "/internal/api-version"); err == nil {
		if bodyBytes, readErr := io.ReadAll(diagResp3.Body); readErr == nil {
			targetAPIVersion = string(bodyBytes)
		}
		diagResp3.Body.Close()
	}

	resp.Data = fiber.Map{
		"internal_service": configK8sServiceURL,
		"k8s": fiber.Map{
			"namespace":    myCFG.Retrievek8sNameSpace(),
			"service_name": myCFG.Retrievek8sServiceName(),
		},
		"target_api_info": fiber.Map{
			"version": targetAPIVersion,
			"status": fiber.Map{
				"online": targetOnline,
				"ready":  targetReady,
			},
		},
		"request_info": fiber.Map{
			"client_ip":  c.IP(),
			"user_agent": c.Get(fiber.HeaderUserAgent),
		},
	}

	return c.JSON(resp)
}

// ProxyRevert godoc
// @Summary revert proxy request to backend
// @Description revert proxy request to backend
// @Tags internal
// @Produce application/json
// @Param use_version query string false "Enable versioned API path using internal api-version"
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /internal/admin/proxy/{routes} [get]
func (r *internalHandlerImpl) ProxyRevert(c *fiber.Ctx) (err error) {
	wildcardPath := c.Params("*")

	configK8sServiceURL := myCFG.RetrieveK8sServiceURL()

	targetURL := configK8sServiceURL

	var b bool
	if q := c.Query("use_version"); q != "" { // only parse when the value is not empty, default to false
		if b, err = strconv.ParseBool(q); err != nil {
			resp := myMDL.Default400Resp.Copy()
			resp.SetErrorAndMessage(err.Error())
			return c.Status(resp.Status).JSON(resp)
		}
	}

	if b {
		targetAPIVersion := "UnKnOwN"
		client := &http.Client{Timeout: 800 * time.Millisecond}
		if diagResp3, err := client.Get(configK8sServiceURL + "/internal/api-version"); err == nil {
			if bodyBytes, readErr := io.ReadAll(diagResp3.Body); readErr == nil {
				targetAPIVersion = string(bodyBytes)
			}
			diagResp3.Body.Close()
		}
		targetURL = targetURL + "/api/" + targetAPIVersion + "/" + wildcardPath
	} else {
		targetURL = targetURL + "/" + wildcardPath
	}

	// Append existing query parameters except "use_version"
	if queryString := string(c.Request().URI().QueryString()); queryString != "" {
		if values, err := url.ParseQuery(queryString); err == nil {
			values.Del("use_version")
			if len(values) > 0 {
				targetURL = targetURL + "?" + values.Encode()
			}
		}
	}

	// Set to private cluster loopback or internal subnet
	c.Request().Header.Set(fiber.HeaderXForwardedFor, "127.0.0.1")
	c.Request().Header.Set("X-Real-IP", "127.0.0.1")

	if err := proxy.Do(c, targetURL); err != nil {
		return c.Status(502).SendString("Config API Service Unreachable")
	}
	return nil
}

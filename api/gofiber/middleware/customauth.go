package middleware

import (
	myERR "gofiber/errors"
	"gofiber/model"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	sessionKeyInContext    = "player_session"
	tarSessionKeyInContext = "tar_ses_bs"

	bearerPrefix = "Bearer "
)

// CustomOAuth2Protected Protected protect routes
func CustomOAuth2Protected(c *fiber.Ctx) (err error) {

	resp := model.Default401Resp.Copy()

	// h := strings.Join(c.GetReqHeaders()["Authorization"], ",")
	// //TODO: this non standard Authorization value doesn't start with "Bearer "
	// a := strings.TrimPrefix(h, "Bearer ")
	// if len(a) <= 3 {
	// 	resp.SetErrorAndMessage(myERR.NoAuth.Error())
	// 	return c.Status(resp.Status).JSON(resp)
	// }

	// Extract the Bearer token from the Authorization header
	var a string
	if a, err = ExtractBearerToken(c); err != nil {
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	// A separately configured static token is accepted only on registered admin routes.
	if isRegisteredAdminRequest(c) && acceptsAdminAPIToken("/admin", a) {
		c.Locals(adminAPITokenLocalKey, true)
		insertSession(c, model.Session{})
		return c.Next()
	}

	var ses model.Session
	if ses, err = model.RetrievePlayerCache(a); err != nil {
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	// pass player_Session fetched from Session
	insertSession(c, ses) //c.Locals("player_session", ses)
	return c.Next()
}

// ExtractBearerToken reads the Authorization header and returns the Bearer token (without the prefix).
func ExtractBearerToken(c *fiber.Ctx) (string, error) {
	// GetReqHeaders() returns map[string][]string
	authHeaders := c.GetReqHeaders()["Authorization"]
	if len(authHeaders) == 0 {
		return "", myERR.ErrMissingAuthHeader
	}

	// If there are multiple headers, join them (rare, but allowed)
	authHeader := strings.Join(authHeaders, ",")

	// if !strings.HasPrefix(authHeader, bearerPrefix) {
	// 	return "", myERR.ErrInvalidAuthHeader
	// }

	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if len(token) <= 3 {
		return "", myERR.ErrInvalidAuthHeader
	}

	return token, nil
}

func ExtractSkipAuthorization(c *fiber.Ctx) string {
	return c.Get("Remote-Authorization-Skip")
}

//// EndPlatformPlaID requires path end with :platform/:pla_id to work
//func EndPlatformPlaID(c *fiber.Ctx) (err error) {
//
//	resp := model.Default400Resp.Copy()
//
//	var pid string
//	if pid, err = url.QueryUnescape(c.Params("pla_id")); err != nil {
//		resp.AddErrorAsMessages(err)
//		return c.Status(resp.Status).JSON(resp)
//	}
//
//	var pla string
//	if pla, err = url.QueryUnescape(c.Params("platform")); err != nil {
//		resp.AddErrorAsMessages(err)
//		return c.Status(resp.Status).JSON(resp)
//	}
//
//	targetSB := model.NewSessionBase(pla, pid)
//
//	// valid input, get error msgs according
//	if errMsgs := targetSB.Validate(); len(errMsgs) > 0 {
//		resp.Error = "Validation Error"
//		resp.Messages = errMsgs
//		return c.Status(resp.Status).JSON(resp)
//	}
//
//	insertTargetSessionBase(c, targetSB)
//
//	return c.Next()
//}

// EndPlatformPlaID requires path end with :platform/:platform_id to work
func EndPlatformPlaID(c *fiber.Ctx) (err error) {

	resp := model.Default400Resp.Copy()

	// from path, we retrieve for this request targetPlayer
	paths := strings.Split(c.Path(), "/")
	length := len(paths)
	if length < 2 {
		resp.AddErrorAsMessages(myERR.ErrSelfPrivilegePathMissing)
		return c.Status(resp.Status).JSON(resp)
	}

	var platformID string
	if platformID, err = url.QueryUnescape(paths[length-1]); err != nil {
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}

	var platform string
	if platform, err = url.QueryUnescape(paths[length-2]); err != nil {
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}

	targetSB := model.NewSessionBase(platform, platformID)

	// valid input, get error msgs according
	if errMsgs := targetSB.Validate(); len(errMsgs) > 0 {
		resp.Error = "Validation Error"
		resp.Messages = errMsgs
		return c.Status(resp.Status).JSON(resp)
	}

	insertTargetSessionBase(c, targetSB)
	return c.Next()
}

// insertSession is a local usage only func
func insertSession(c *fiber.Ctx, session model.Session) {
	c.Locals(sessionKeyInContext, session)
}

func RetrieveSession(c *fiber.Ctx) model.Session {
	return c.Locals(sessionKeyInContext).(model.Session)
}

func insertTargetSessionBase(c *fiber.Ctx, sb model.SessionBase) {
	c.Locals(tarSessionKeyInContext, sb)
}

func RetrieveTargetSessionBase(c *fiber.Ctx) model.SessionBase {
	return c.Locals(tarSessionKeyInContext).(model.SessionBase)
}

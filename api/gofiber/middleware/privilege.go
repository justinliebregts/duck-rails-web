package middleware

import (
	myERR "gofiber/errors"
	"gofiber/model"

	"github.com/gofiber/fiber/v2"
)

func checkAdminPrivilege(playerUUID string) (err error, extraLog string) {

	var adminPrivilege model.AdminPrivilege
	if adminPrivilege, err = model.RetrieveAdminPrivilege(playerUUID); err != nil {
		extraLog = myERR.NoAdminPrivilegeFound.Error()
	} else if adminPrivilege.DeletedAt.Valid && adminPrivilege.DeletedAt.Time.Before(model.CurrentTimestamp()) {
		// AdminPrivilege Found but Expired
		err = myERR.ErrPrivilegeExpired
	}

	return
}

const (
	isAdminLocalKy = "is_admin"
)

// CheckAdminPrivilegeWithSession is layering on checkAdminPrivilege
// CheckAdminPrivilegeWithSession is NOT a standard middleware to chain with handler!
func CheckAdminPrivilegeWithSession(c *fiber.Ctx) (isAdmin bool, extraLog string, err error) {
	if verified, _ := c.Locals(adminAPITokenLocalKey).(bool); verified {
		c.Locals(isAdminLocalKy, true)
		return true, "", nil
	}
	ses := RetrieveSession(c)

	if err, extraLog = checkAdminPrivilege(ses.ComposedID()); err == nil {
		c.Locals(isAdminLocalKy, true)
		isAdmin = true
	}

	return
}

// AdminPrivilegeCheckMiddleware is a middleware to chain with handler
func AdminPrivilegeCheckMiddleware(c *fiber.Ctx) error {
	isAdmin, extraLog, err := CheckAdminPrivilegeWithSession(c)
	if isAdmin {
		return c.Next() // Continue processing if admin
	}

	// err found; stop processing; give response
	resp := model.Default403Resp.Copy()
	resp.SetErrorAndMessage(err.Error())
	if len(extraLog) > 0 {
		resp.AddMessages(extraLog)
	}

	return c.Status(resp.Status).JSON(resp)
}

// OwnerOrAdminWithTargetSessionBaseCheckMiddleware is a middleware to chain with handler
// OwnerOrAdminWithTargetSessionBaseCheckMiddleware is after EndPlatformPlaID
func OwnerOrAdminWithTargetSessionBaseCheckMiddleware(c *fiber.Ctx) error {
	if isAdmin, _, err := CheckAdminPrivilegeWithSession(c); err == nil && isAdmin {
		// is Admin; PASS
		return c.Next()
	}

	ses := RetrieveSession(c)
	tarSB := RetrieveTargetSessionBase(c)
	if tarSB.ComposedID() == ses.ComposedID() {
		// is Owner; PASS
		return c.Next()
	}

	// Both Check failed; no Privilege at all
	resp := model.Default403Resp.Copy()
	resp.AddMessages(myERR.ErrSelfPrivilegeNotOwner.Error())
	return c.Status(resp.Status).JSON(resp)
}

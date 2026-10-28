package handler

import (
	"errors"
	myDOM "gofiber/domain"
	"gofiber/model"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

var (
	// use a single instance of Validate, it caches struct info
	validate = myDOM.GetValidator()
)

var (
	devSession            = model.Session{SessionBase: model.SessionBase{PlayerID: "3724467550970486", Platform: "OCULUS"}}
	devSessionAccessToken = "Xe94YkdkD9ejWG5rp42nugeHnEr"
)

func CreateDevSession() {
	log.Println("Creating DEV only Session")
	// create/update Player_Online_Session into Cache
	ses := devSession
	ses.AutoFill()
	ses.AccessToken = devSessionAccessToken
	_ = ses.SaveToCacheForDuration(356 * 24 * time.Hour)
}

func GiveDevSessionAdminPrivilege() {
	grUUID := devSession.PlayerID
	if _, err := model.RetrieveAdminPrivilege(grUUID); err == nil {
		return
	}

	log.Println("Giving DevSession Admin Privilege")
	pri := model.AdminPrivilege{PlayerID: grUUID}
	pri.Save()
}

func LoadCustomErrorHandler(c *fiber.Ctx, err error) error {
	var e *fiber.Error
	if errors.As(err, &e) {
		c.Status(e.Code)

		switch e.Code {
		case fiber.StatusBadRequest:
			return c.JSON(model.Default400Resp)
		case fiber.StatusUnauthorized:
			return c.JSON(model.Default401Resp)
		case fiber.StatusNotFound:
			return c.JSON(model.Default404Resp)
		case fiber.StatusMethodNotAllowed:
			return c.JSON(model.Default405Resp)
		case fiber.StatusServiceUnavailable:
			return c.JSON(model.Default503Resp)
		}
	}

	return c.SendString(err.Error())
}

package handler

import (
	myERR "gofiber/errors"
	"gofiber/middleware"
	"gofiber/model"
	mySRV "gofiber/service"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type PlayerHandler interface {
	PlayerSocialAccountHandler(c *fiber.Ctx) error
	ListAttestationRecordHandler(c *fiber.Ctx) (err error)

	GetOculusOrgScopedIDHandler(c *fiber.Ctx) error
	PlayerOculusAttestationToggleBanHandler(c *fiber.Ctx) error
	ListAttestationDeviceBanRecordHandler(c *fiber.Ctx) error
	CheckRemoteAttestationBanStatus(c *fiber.Ctx) error

	PlayerAdminListHandler(c *fiber.Ctx) error
}

func SetPlayerConfig(freqInSec int, durInSec int) {
	model.HeartbeatFrequency = time.Second * time.Duration(freqInSec)
	model.HeartbeatLifeCycle = time.Second * time.Duration(durInSec)

	log.Printf("Heartbeat freq:%v lifeCycle:%v", model.HeartbeatFrequency, model.HeartbeatLifeCycle)
}

type PlayerSearchForm struct {
	PlayerID   string `form:"player_id" json:"player_id" validate:"required_without=PlayerName,omitempty"`
	PlayerName string `form:"player_name" json:"player_name" validate:"required_without=PlayerID,omitempty,min=1"`
}

func (p *PlayerSearchForm) Validate() (errMsgs []string) {
	if err := validate.Struct(p); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

// PlayerPostHeartbeat godoc
// @Summary
// @Description
// @Tags player
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param form formData model.HeartbeatForm true "Heartbeat Form"
// @Success 200 {object} model.ResponseGeneral{status=int,data=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /player/heartbeat [post]
func PlayerPostHeartbeat(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	//Parse POST Input
	usrInput := new(model.HeartbeatForm)
	if err = c.BodyParser(usrInput); err != nil {
		resp = model.Default503Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	// valid input, get error msgs according
	if errMsgs := usrInput.Validate(); len(errMsgs) > 0 {
		resp = model.Default400Resp.Copy()
		resp.Error = "Validation Error"
		resp.Messages = errMsgs
		return c.Status(resp.Status).JSON(resp)
	}

	ses := middleware.RetrieveSession(c)

	if _, err = model.CacheSavePlayerState(&ses.SessionBase, *usrInput); err != nil {
		resp = model.Default503Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	resp = model.Default200Resp.Copy()
	resp.AddMessages("Aliveness reported")
	return c.JSON(resp)
}

// PlayerPostTerminateHeartbeat godoc
// @Summary
// @Description
// @Tags player
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param form formData model.HeartbeatForm true "Heartbeat Form"
// @Success 200 {object} model.ResponseGeneral{status=int,data=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /player/admin/cease/{player_uuid} [post]
func PlayerPostTerminateHeartbeat(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	pid := c.Params("player_uuid", "0")
	if len(pid) < 1 {
		resp = model.Default400Resp.Copy()
		resp.AddErrorAsMessages(myERR.NoPlayerMatch)
		return c.Status(resp.Status).JSON(resp)
	}

	targetSes := model.FormSessionBase(pid)

	if err = model.CacheRemovePlayerState(&targetSes); err != nil {
		resp = model.Default503Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	resp = model.Default200Resp.Copy()
	resp.AddMessages(targetSes.ComposedID() + " Terminated")
	return c.JSON(resp)
}

// PlayerListAlive godoc
// @Summary list player that set Heartbeat
// @Description
// @Tags player
// @Accept */*
// @Produce json
// @Success 200 {object} model.ResponseGeneral{status=int,data=[]string}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /player/online [get]
func PlayerListAlive(c *fiber.Ctx) (err error) {
	resp := model.Default200Resp.Copy()
	resp.Data = (&model.HeartbeatForm{}).All()
	resp.AddMessages("Online players fetched")
	return c.JSON(resp)
}

// SharedAccountToSessionModule replace SharedAccountToSession with injecting repo/service for module
func SharedAccountToSessionModule(c *fiber.Ctx, usrInput *model.PlayerRegistrationForm, playerSvc mySRV.Player) (err error) {
	var resp model.ResponseGeneral
	scfPtr := &usrInput.SessionCreateForm

	////Verify Account with remote API
	//if errMsgs := playerSvc.ValidateWithDevicePlatform(scfPtr); len(errMsgs) > 0 {
	//	resp = model.Default455Resp.Copy()
	//	resp.Messages = errMsgs
	//	return c.Status(resp.Status).JSON(resp)
	//}

	// Create report and block if AppHashRecord not match
	if err = mySRV.HashVersionCheckAndReport(scfPtr); err != nil {
		resp = model.Default403Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}

	if playerSvc.IsDeviceBlocked(scfPtr) || usrInput.IsDeviceNameBlocked() {
		resp = model.Default403Resp.Copy()
		resp.AddErrorAsMessages(myERR.DeviceBanned)
		return c.Status(resp.Status).JSON(resp)
	}

	//Create or Retrieve PlayerRecord
	var playerPtr *model.Player
	if playerPtr, _, err = playerSvc.SyncPlayerRecord(usrInput); err != nil {
		resp = model.Default503Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}

	// Check if player Banned from Network services
	var banned bool
	if banned, err = playerPtr.IsGameBanned(); err != nil {
		resp = model.Default503Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	} else if banned {
		resp = model.Default403Resp.Copy()
		resp.AddErrorAsMessages(myERR.Banned)
		return c.Status(resp.Status).JSON(resp)
	}

	// create/update Player_Online_Session into Cache
	var session model.Session
	if session, err = model.CreatePlayerCache(usrInput.PlayerID, usrInput.Platform); err != nil {
		resp = model.Default503Resp.Copy()
		resp.Error = "Cache Error"
		resp.Message = err.Error()
		return c.Status(resp.Status).JSON(resp)
	}

	//Save display name in Cache; optional; non-blocking if failed
	_ = model.TrySaveDisplayNameToCache(&session.SessionBase, playerPtr.Name)

	//TODO: session exist check and create?

	resp = model.Default200Resp.Copy()
	resp.AddMessages("Player updated")

	resp.Data = session.FormResponseForRegister(playerPtr.Name)
	return c.JSON(resp)
}

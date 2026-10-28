package handler

import (
	"errors"
	"gofiber/model"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// PlayerGetAdminHandler godoc
// @Summary
// @Description
// @Tags player
// @Accept application/json
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param player_id query string false "player ID"
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 404 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /player/admin/get [get]
// @Router /player/admin/get/{player_id} [get]
func PlayerGetAdminHandler(c *fiber.Ctx) (err error) {
	resp := model.Default200Resp.Copy()

	// query
	composedPlayerID := c.Query("player_id")
	//platform := c.Query("platform", "OCULUS")
	var player model.Player
	if player, err = model.RetrieveFromDB(composedPlayerID); err != nil {
		resp = model.Default500Resp.Copy()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp = model.Default404Resp.Copy()
		}
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	resp.Data = player.ToMap()
	return c.JSON(resp)
}

// PlayerSearchAdminHandler godoc
// @Summary
// @Description
// @Tags player
// @Accept application/json
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param player_id query string false "player ID"
// @Param player_name query string false "player Name"
// @Success 200 {object} model.ResponseGeneral{status=int,data=[]model.Player}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /player/admin/search [get]
// @Router /player/admin/search-by-id-admin [get]
// @Router /player/admin/search-by-name-admin [get]
func PlayerSearchAdminHandler(c *fiber.Ctx) (err error) {
	resp := model.Default200Resp.Copy()

	psf := PlayerSearchForm{PlayerName: c.Query("player_name"), PlayerID: c.Query("player_id")}
	if errMsgs := psf.Validate(); len(errMsgs) > 0 {
		resp = model.Default400Resp.Copy()
		resp.Error = "Validation Error"
		resp.Messages = errMsgs
		return c.Status(resp.Status).JSON(resp)
	}

	var players []model.Player
	//determine ById or ByName
	if len(psf.PlayerID) > 0 {
		if players, err = model.SearchByID(psf.PlayerID); err != nil {
			resp = model.Default503Resp.Copy()
			resp.Error = "SearchByID Error"
			resp.Message = err.Error()
			return c.Status(resp.Status).JSON(resp)
		}
		resp.Message = "SearchByID"
	} else {
		if players, err = model.SearchByName(psf.PlayerName); err != nil {
			resp = model.Default503Resp.Copy()
			resp.Error = "SearchByName Error"
			resp.Message = err.Error()
			return c.Status(resp.Status).JSON(resp)
		}
		resp.Message = "SearchByName"
	}

	resp.Data = players
	return c.JSON(resp)
}

// PlayerGetAllAdminHandler godoc
// @Summary
// @Description
// @Tags player
// @Accept */*
// @Produce json
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /player/admin/all [get]
func PlayerGetAllAdminHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	var players []model.Player

	if players, err = model.GetAllPlayers(); err != nil {
		resp = model.Default500Resp.Copy()
		resp.Error = err.Error()
		resp.Message = "Failed to get all players"
		return c.Status(resp.Status).JSON(resp)
	}
	resp = model.Default200Resp.Copy()
	resp.Message = "Players found"
	if len(players) == 0 {
		players = []model.Player{}
		resp.Message = "No players found"
	}
	resp.Data = players
	return c.JSON(resp)
}

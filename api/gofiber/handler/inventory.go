package handler

import (
	"errors"
	"fmt"
	myERR "gofiber/errors"
	"gofiber/middleware"
	"gofiber/model"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var ()

//// AllInventoryHandler godoc
//// @Summary deprecated v1, using new v2
//// @Description
//// @Tags inventory
//// @Accept */*
//// @Produce json
//// @Security ApiKeyNonStandardAuth
//// @Param dev_owned path int false "dev_owned"
//// @Success 200 {object} model.ResponseGeneral{status=int}
//// @Failure 400 {object} model.ResponseGeneral{status=int}
//// @Failure 500 {object} model.ResponseGeneral{status=int}
//// @Failure 503 {object} model.ResponseGeneral{status=int}
//func AllInventoryHandler(c *fiber.Ctx) (err error) {
//	var resp model.ResponseGeneral
//	ses := middleware.RetrieveSession(c)
//
//	// Params: /inventory/all/{dev_owned}
//	devOwned := c.Params("dev_owned", "0") != "0"
//
//	var items []model.GetItemsResponse
//	if items, err = model.GetItems(ses.ComposedID(), devOwned); err != nil {
//		resp = model.Default500Resp.Copy()
//		resp.Error = err.Error()
//		return c.Status(resp.Status).JSON(resp)
//	}
//
//	resp = model.Default200Resp.Copy()
//	resp.Data = items
//	resp.Message = "Items retrieved successfully"
//
//	return c.JSON(resp)
//}

// DeleteAdminInventoryHandler godoc
// @Summary
// @Description
// @Tags inventory
// @Accept */*
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param transaction_id path int true "Transaction ID"
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /inventory/admin/delete/{transaction_id} [post]
func DeleteAdminInventoryHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	var transactionID uint64
	if transactionID, err = strconv.ParseUint(c.Params("transaction_id"), 10, 64); err != nil {
		resp = model.Default400Resp.Copy()
		resp.Error = err.Error()
		resp.AddMessages("Transaction ID is required")
		return c.Status(fiber.StatusBadRequest).JSON(resp)
	}

	var isDeletedNum int64
	if isDeletedNum, err = strconv.ParseInt(c.FormValue("is_deleted"), 10, 64); err != nil {
		resp = model.Default400Resp.Copy()
		resp.Error = err.Error()
		resp.AddMessages("Is Deleted is required")
		return c.Status(fiber.StatusBadRequest).JSON(resp)
	}

	isDeleted := false
	if isDeletedNum > 0 {
		isDeleted = true
	}

	if err = model.DeleteTransactionAdmin(uint(transactionID), isDeleted); err != nil {
		resp = model.Default500Resp.Copy()
		resp.Error = err.Error()
		resp.AddMessages("Failed to delete item")
		return c.Status(resp.Status).JSON(resp)
	}

	resp = model.Default200Resp.Copy()
	if isDeleted {
		resp.AddMessages(fmt.Sprintf("Transaction %d successfully deleted", transactionID))
	} else {
		resp.AddMessages(fmt.Sprintf("Transaction %d successfully restored", transactionID))
	}
	resp.Data = transactionID
	return c.JSON(resp)
}

// AllInventoryHandler godoc
// @Summary List owned avatar and weapon skins
// @Tags inventory
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param dev_owned path int false "include dev-owned items when non-zero"
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /inventory/all [get]
// @Router /inventory/all/{dev_owned} [get]
func AllInventoryHandler(c *fiber.Ctx) error {
	var resp model.ResponseGeneral
	ses := middleware.RetrieveSession(c)
	devOwned := c.Params("dev_owned", c.Query("dev_owned", "0")) != "0"
	items, err := model.GetItems(ses.ComposedID(), devOwned)
	if err != nil {
		resp = model.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	resp = model.Default200Resp.Copy()
	resp.Data = items
	resp.AddMessages("Items retrieved successfully")
	return c.JSON(resp)
}

type BuyItemForm struct {
	ItemID          uint   `form:"item_id" json:"item_id" validate:"required"`
	Platform        string `form:"platform" json:"platform" validate:"required"`
	PlatformAccount string `form:"platform_account" json:"platform_account"`
	PlatformTxnID   string `form:"platform_txn_id" json:"platform_txn_id" validate:"required"`
	PlatformTxnTime string `form:"platform_txn_time" json:"platform_txn_time"`
}

func (p *BuyItemForm) Validate() (errMsgs []string) {
	if err := validate.Struct(p); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

// BuyInventoryHandler godoc
// @Summary Record a platform purchase
// @Description Inserts an idempotent transaction. Coin balances are not checked or changed.
// @Tags inventory
// @Accept json
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param form body BuyItemForm true "platform purchase"
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /inventory/buy [post]
func BuyInventoryHandler(c *fiber.Ctx) error {
	var resp model.ResponseGeneral
	ses := middleware.RetrieveSession(c)
	form := new(BuyItemForm)
	if err := c.BodyParser(form); err != nil {
		resp = model.Default400Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	if errMsgs := form.Validate(); len(errMsgs) > 0 {
		resp = model.Default400Resp.Copy()
		resp.Error = "Validation Error"
		resp.Messages = errMsgs
		return c.Status(resp.Status).JSON(resp)
	}

	recorded, err := model.IsPlatformTxnRecorded(form.PlatformTxnID)
	if err != nil {
		resp = model.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	if recorded {
		resp = model.Default200Resp.Copy()
		resp.AddMessages("Purchase already recorded")
		resp.Data = 0
		return c.JSON(resp)
	}

	err = model.RecordPurchase(ses.ComposedID(), form.ItemID, form.Platform, form.PlatformAccount, form.PlatformTxnID, form.PlatformTxnTime)
	if err != nil {
		if errors.Is(err, myERR.ErrCannotOwnMultiple) {
			resp = model.Default200Resp.Copy()
			resp.AddMessages("Item already owned")
			resp.Data = 0
			return c.JSON(resp)
		}
		resp = model.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	resp = model.Default200Resp.Copy()
	resp.AddMessages("Purchase recorded")
	resp.Data = form.ItemID
	return c.JSON(resp)
}

// AllAdminInventoryHandler godoc
// @Summary List a player's skin transactions
// @Tags inventory
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param player_id path string true "player id"
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /inventory/admin/all/{player_id} [get]
func AllAdminInventoryHandler(c *fiber.Ctx) error {
	var resp model.ResponseGeneral
	playerID := c.Params("player_id")
	if playerID == "" {
		resp = model.Default400Resp.Copy()
		resp.AddMessages("Player ID is required")
		return c.Status(resp.Status).JSON(resp)
	}
	rows, err := model.GetItemsAdmin(playerID)
	if err != nil {
		resp = model.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	resp = model.Default200Resp.Copy()
	resp.Data = rows
	resp.AddMessages("Transactions found")
	return c.JSON(resp)
}

type GiveItemForm struct {
	PlayerID string `form:"player_id" json:"player_id" validate:"required"`
	ItemID   uint   `form:"item_id" json:"item_id" validate:"required"`
}

// GiveItemInventoryHandler godoc
// @Summary Give an avatar or weapon skin
// @Description Inserts an ADMIN transaction with no coin change.
// @Tags inventory
// @Accept json
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param form body GiveItemForm true "player and item"
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /inventory/admin/give-item [post]
func GiveItemInventoryHandler(c *fiber.Ctx) error {
	var resp model.ResponseGeneral
	form := new(GiveItemForm)
	if err := c.BodyParser(form); err != nil {
		resp = model.Default400Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	if errMsgs := validateGive(form); len(errMsgs) > 0 {
		resp = model.Default400Resp.Copy()
		resp.Error = "Validation Error"
		resp.Messages = errMsgs
		return c.Status(resp.Status).JSON(resp)
	}
	if _, err := model.GiveItemAdmin(form.PlayerID, form.ItemID); err != nil {
		resp = model.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		resp.AddMessages("Failed to give item")
		return c.Status(resp.Status).JSON(resp)
	}
	resp = model.Default200Resp.Copy()
	resp.AddMessages("Item added to player's inventory")
	resp.Data = form.ItemID
	return c.JSON(resp)
}

func validateGive(form *GiveItemForm) []string {
	if err := validate.Struct(form); err != nil {
		var msgs []string
		for _, er := range err.(validator.ValidationErrors) {
			msgs = append(msgs, er.Error())
		}
		return msgs
	}
	return nil
}

// DeleteAllAdminInventoryHandler godoc
// @Summary
// @Description
// @Tags inventory
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param player_id formData string true "Player ID"
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /inventory/admin/delete-all [post]
func DeleteAllAdminInventoryHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral
	playerID := c.FormValue("player_id")
	if playerID == "" {
		resp = model.Default400Resp.Copy()
		resp.Message = "Player ID is required"
		return c.Status(fiber.StatusBadRequest).JSON(resp)
	}

	if err = model.DeleteAllTransactionsAdmin(playerID); err != nil {
		resp = model.Default500Resp.Copy()
		resp.Error = err.Error()
		resp.Message = "Failed to delete all items"
		return c.Status(resp.Status).JSON(resp)
	}

	resp = model.Default200Resp.Copy()
	resp.Message = "All transactions successfully deleted"
	return c.JSON(resp)
}

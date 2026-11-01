package duckz

import (
	"errors"
	"log"

	myERR "gofiber/errors"
	myMW "gofiber/middleware"
	myMDL "gofiber/model"
	dzSRV "gofiber/service/duckz"

	"github.com/gofiber/fiber/v2"
)

type EconomyHandler interface {
	Wallet(c *fiber.Ctx) error
	Store(c *fiber.Ctx) error
	Deposit(c *fiber.Ctx) error
	Buy(c *fiber.Ctx) error
	GrantIAP(c *fiber.Ctx) error
	SaveLoadout(c *fiber.Ctx) error
	AdminWallet(c *fiber.Ctx) error
	AdminAdjust(c *fiber.Ctx) error
	AdminTransactions(c *fiber.Ctx) error
	AdminCheck(c *fiber.Ctx) error
}

type economyHandlerImpl struct {
	svc dzSRV.EconomyService
}

func NewEconomyHandler(svc dzSRV.EconomyService) EconomyHandler {
	return &economyHandlerImpl{svc: svc}
}

func errorResp(err error) myMDL.ResponseGeneral {
	var resp myMDL.ResponseGeneral
	switch {
	case errors.Is(err, myERR.ErrDuckzInsufficientCoins):
		resp = myMDL.Default402Resp.Copy()
	case errors.Is(err, myERR.ErrDuckzDepositCapReached):
		resp = myMDL.Default429Resp.Copy()
	case errors.Is(err, myERR.ErrDuckzEntitlementNotOwned),
		errors.Is(err, myERR.ErrDuckzRequestIDOtherPlayer),
		errors.Is(err, myERR.ErrDuckzPlatformTxnIDUsed),
		errors.Is(err, myERR.ErrDuckzItemNotOwned):
		resp = myMDL.Default403Resp.Copy()
	case errors.Is(err, myERR.ErrDuckzInvalidRequestID),
		errors.Is(err, myERR.ErrDuckzInvalidBars),
		errors.Is(err, myERR.ErrDuckzUnknownItem),
		errors.Is(err, myERR.ErrDuckzItemNotForSale),
		errors.Is(err, myERR.ErrDuckzUnknownSku),
		errors.Is(err, myERR.ErrDuckzInvalidPlatformTxnID),
		errors.Is(err, myERR.ErrDuckzUnsupportedPlatform),
		errors.Is(err, myERR.ErrDuckzPlayerNotFound),
		errors.Is(err, myERR.ErrDuckzInvalidAmount):
		resp = myMDL.Default400Resp.Copy()
	default:
		resp = myMDL.Default500Resp.Copy()
	}
	resp.AddErrorAsMessages(err)
	return resp
}

// Wallet godoc
// @Summary Duckz wallet: coin balance, owned IAP skus, purchased store item keys
// @Tags duckz
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Success 200 {object} model.ResponseGeneral{data=duckz.WalletResult}
// @Router /wallet [get]
func (h *economyHandlerImpl) Wallet(c *fiber.Ctx) error {
	ses := myMW.RetrieveSession(c)
	result, err := h.svc.Wallet(ses.ComposedID())
	if err != nil {
		log.Printf("[duckz.wallet] error player=%s: %v", ses.ComposedID(), err)
		resp := errorResp(err)
		return c.Status(resp.Status).JSON(resp)
	}
	resp := myMDL.Default200Resp.Copy()
	resp.Data = result
	return c.Status(resp.Status).JSON(resp)
}

// Store godoc
// @Summary Duckz coin store prices from duckz_economy.json
// @Tags duckz
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Success 200 {object} model.ResponseGeneral{data=duckz.StoreListResult}
// @Router /store [get]
func (h *economyHandlerImpl) Store(c *fiber.Ctx) error {
	resp := myMDL.Default200Resp.Copy()
	resp.Data = h.svc.Store()
	return c.Status(resp.Status).JSON(resp)
}

// Deposit godoc
// @Summary Convert deposited gold bars to duck coins (server computes coins, capped per hour, idempotent on request_id)
// @Tags duckz
// @Accept json
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param body body duckz.DepositRequest true "Deposit"
// @Success 200 {object} model.ResponseGeneral{data=duckz.DepositResult}
// @Failure 429 {object} model.ResponseGeneral{status=int}
// @Router /coins/deposit [post]
func (h *economyHandlerImpl) Deposit(c *fiber.Ctx) error {
	var req dzSRV.DepositRequest
	if err := c.BodyParser(&req); err != nil {
		resp := myMDL.Default400Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}
	ses := myMW.RetrieveSession(c)
	result, err := h.svc.Deposit(ses.ComposedID(), req, myMDL.CurrentTimestamp())
	if err != nil {
		resp := errorResp(err)
		return c.Status(resp.Status).JSON(resp)
	}
	resp := myMDL.Default200Resp.Copy()
	resp.Data = result
	return c.Status(resp.Status).JSON(resp)
}

// Buy godoc
// @Summary Buy a static store item with duck coins (server price, idempotent on request_id)
// @Tags duckz
// @Accept json
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param body body duckz.BuyRequest true "Buy"
// @Success 200 {object} model.ResponseGeneral{data=duckz.BuyResult}
// @Failure 402 {object} model.ResponseGeneral{data=duckz.BuyResult}
// @Router /store/buy [post]
func (h *economyHandlerImpl) Buy(c *fiber.Ctx) error {
	var req dzSRV.BuyRequest
	if err := c.BodyParser(&req); err != nil {
		resp := myMDL.Default400Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}
	ses := myMW.RetrieveSession(c)
	result, err := h.svc.Buy(ses.ComposedID(), req, myMDL.CurrentTimestamp())
	if err != nil {
		resp := errorResp(err)
		if result != nil {
			resp.Data = result
		}
		return c.Status(resp.Status).JSON(resp)
	}
	resp := myMDL.Default200Resp.Copy()
	resp.Data = result
	if result.AlreadyOwned {
		resp.AddMessages("already owned")
	}
	return c.Status(resp.Status).JSON(resp)
}

// GrantIAP godoc
// @Summary Record a durable Meta IAP after verifying ownership with Meta S2S
// @Tags duckz
// @Accept json
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param body body duckz.IAPGrantRequest true "Grant"
// @Success 200 {object} model.ResponseGeneral{data=duckz.IAPGrantResult}
// @Router /iap/grant [post]
func (h *economyHandlerImpl) GrantIAP(c *fiber.Ctx) error {
	var req dzSRV.IAPGrantRequest
	if err := c.BodyParser(&req); err != nil {
		resp := myMDL.Default400Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}
	ses := myMW.RetrieveSession(c)
	result, err := h.svc.GrantIAP(ses.ComposedID(), ses.Platform, req, myMDL.CurrentTimestamp())
	if err != nil {
		resp := errorResp(err)
		return c.Status(resp.Status).JSON(resp)
	}
	resp := myMDL.Default200Resp.Copy()
	resp.Data = result
	return c.Status(resp.Status).JSON(resp)
}

// SaveLoadout godoc
// @Summary Save the equipped avatar ("" for the default duck); the key must be owned
// @Tags duckz
// @Accept json
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param body body duckz.LoadoutRequest true "Loadout"
// @Success 200 {object} model.ResponseGeneral{data=duckz.LoadoutResult}
// @Failure 403 {object} model.ResponseGeneral{status=int}
// @Router /loadout [post]
func (h *economyHandlerImpl) SaveLoadout(c *fiber.Ctx) error {
	var req dzSRV.LoadoutRequest
	if err := c.BodyParser(&req); err != nil {
		resp := myMDL.Default400Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}
	ses := myMW.RetrieveSession(c)
	result, err := h.svc.SaveLoadout(ses.ComposedID(), req, myMDL.CurrentTimestamp())
	if err != nil {
		resp := errorResp(err)
		return c.Status(resp.Status).JSON(resp)
	}
	resp := myMDL.Default200Resp.Copy()
	resp.Data = result
	return c.Status(resp.Status).JSON(resp)
}

// AdminWallet godoc
// @Summary Duckz admin: a player's coin balance, owned items and recent admin coin adjustments
// @Tags duckz
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param player_id path string true "Player ID"
// @Success 200 {object} model.ResponseGeneral{data=duckz.AdminWalletResult}
// @Router /admin/wallet/{player_id} [get]
func (h *economyHandlerImpl) AdminWallet(c *fiber.Ctx) error {
	result, err := h.svc.AdminWallet(c.Params("player_id"))
	if err != nil {
		resp := errorResp(err)
		return c.Status(resp.Status).JSON(resp)
	}
	resp := myMDL.Default200Resp.Copy()
	resp.Data = result
	return c.Status(resp.Status).JSON(resp)
}

// AdminAdjust godoc
// @Summary Duckz admin: grant (positive amount) or remove (negative) duck coins
// @Tags duckz
// @Accept json
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param body body duckz.AdminAdjustRequest true "Adjustment"
// @Success 200 {object} model.ResponseGeneral{data=duckz.AdminAdjustResult}
// @Failure 402 {object} model.ResponseGeneral{data=duckz.AdminAdjustResult}
// @Router /admin/coins/adjust [post]
func (h *economyHandlerImpl) AdminAdjust(c *fiber.Ctx) error {
	var req dzSRV.AdminAdjustRequest
	if err := c.BodyParser(&req); err != nil {
		resp := myMDL.Default400Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}
	ses := myMW.RetrieveSession(c)
	result, err := h.svc.AdminAdjust(ses.ComposedID(), req)
	if err != nil {
		resp := errorResp(err)
		if result != nil {
			resp.Data = result
		}
		return c.Status(resp.Status).JSON(resp)
	}
	resp := myMDL.Default200Resp.Copy()
	resp.Data = result
	return c.Status(resp.Status).JSON(resp)
}

// AdminTransactions godoc
// @Summary Duckz admin: a player's latest transactions (coins, store and IAP rows)
// @Tags duckz
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param player_id path string true "Player ID"
// @Success 200 {object} model.ResponseGeneral{data=[]duckz.AdminTransaction}
// @Router /admin/transactions/{player_id} [get]
func (h *economyHandlerImpl) AdminTransactions(c *fiber.Ctx) error {
	rows, err := h.svc.AdminTransactions(c.Params("player_id"))
	if err != nil {
		resp := errorResp(err)
		return c.Status(resp.Status).JSON(resp)
	}
	resp := myMDL.Default200Resp.Copy()
	resp.Data = rows
	return c.Status(resp.Status).JSON(resp)
}

// AdminCheck succeeds only for admins; AdminPrivilegeCheckMiddleware answers 403 for everyone else.
func (h *economyHandlerImpl) AdminCheck(c *fiber.Ctx) error {
	resp := myMDL.Default200Resp.Copy()
	return c.Status(resp.Status).JSON(resp)
}

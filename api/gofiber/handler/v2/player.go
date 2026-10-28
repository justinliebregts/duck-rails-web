package v2

import (
	myERR "gofiber/errors"
	v1 "gofiber/handler"
	"gofiber/internal/service/remote/oculus"
	oculueRmtSRV "gofiber/internal/service/remote/oculus"
	myMW "gofiber/middleware"
	"gofiber/model"
	myREPO "gofiber/repository"
	mySRV "gofiber/service"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type playerV2HandlerImpl struct {
	playerRepo myREPO.PlayerRepository

	playerSvc mySRV.Player
	oculusSvc oculueRmtSRV.Client
}

func NewPlayerV2Handler(playerRepo myREPO.PlayerRepository, playerSvc mySRV.Player, oculusSvc oculueRmtSRV.Client) v1.PlayerHandler {
	return &playerV2HandlerImpl{
		playerRepo: playerRepo,
		playerSvc:  playerSvc,
		oculusSvc:  oculusSvc,
	}
}

// PlayerSocialAccountHandler godoc
// @Summary Register or sign in using a social account; block any nonSocialAccount
// @Description Creates or logs in a Player using a social account, updates DB and player existence cache
// @Tags player
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param registration formData model.PlayerRegistrationForm true "Player Registration Form"
// @Success 200 {object} model.ResponseGeneral{status=int,data=model.Session}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /player/v2/social-account/{platform} [post]
func (p *playerV2HandlerImpl) PlayerSocialAccountHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	var usrInputPtr *model.PlayerRegistrationForm

	// Retrieve the already-parsed struct from the middleware
	val := c.Locals(myMW.VerifiedRegistrationKey)
	if val != nil {
		usrInputPtr = val.(*model.PlayerRegistrationForm)
	}

	if usrInputPtr == nil {
		//Parse POST Input
		usrInputPtr = new(model.PlayerRegistrationForm)
		if err = c.BodyParser(usrInputPtr); err != nil {
			resp = model.Default503Resp.Copy()
			resp.AddErrorAsMessages(err)
			return c.Status(resp.Status).JSON(resp)
		}
	}

	// valid input, get error msgs according
	if errMsgs := usrInputPtr.Validate(); len(errMsgs) > 0 {
		resp = model.Default400Resp.Copy()
		resp.Error = "Validation Error"
		resp.Messages = errMsgs
		return c.Status(resp.Status).JSON(resp)
	}

	usrInputPtr.FillIPs(c)

	if usrInputPtr.Platform == model.PlatformEMail.Name {
		resp = model.Default400Resp.Copy()
		resp.AddErrorAsMessages(myERR.ErrOnlyAvailableForSocialAccounts)
		return c.Status(resp.Status).JSON(resp)
	}

	return v1.SharedAccountToSessionModule(c, usrInputPtr, p.playerSvc)
}

// GetOculusOrgScopedIDHandler godoc
// @Summary      Retrieve Oculus Verified Org-Scoped ID
// @Description  Exchanges a Meta SDK User ID for a verified Oculus Org-Scoped ID
// @Tags         social-account
// @Accept       json
// @Produce      json
// @Param        oculus_id query string true "Oculus User ID"
// @Success      200 {object} model.ResponseGeneral{data=map[string]string}
// @Failure      400 {object} model.ResponseGeneral{status=int}
// @Failure      500 {object} model.ResponseGeneral{status=int}
// @Router       /player/v2/social-account/admin/oculus-org-scoped-id [get]
func (p *playerV2HandlerImpl) GetOculusOrgScopedIDHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	oculusID := c.Query("oculus_id", "")
	if len(oculusID) == 0 {
		resp = model.Default400Resp.Copy()
		resp.AddErrorAsMessages(myERR.ErrMailboxInvalidPlayer)
		return c.Status(resp.Status).JSON(resp)
	}

	var respOrgScopedID model.GetOculusOrgScopedIDResponse
	if respOrgScopedID, err = p.oculusSvc.GetOculusOrgScopedID(oculusID, nil); err != nil {
		resp = model.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	resp = model.Default200Resp.Copy()
	resp.Data = fiber.Map{
		"id":        respOrgScopedID.ID,
		"alias":     respOrgScopedID.Alias,
		"scoped_id": respOrgScopedID.ScopedID,
	}
	return c.Status(resp.Status).JSON(resp)
}

// ListAttestationRecordHandler godoc
// @Summary      List Attestation Records
// @Description  Retrieve attestation logs with optional filtering by Player ID and App Source.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id          query    string  false  "Platform ID"
// @Param        app_source  query    bool    false  "Filter by current app source (default: false)"
// @Success      200 {object} model.ResponseGeneral
// @Failure      400 {object} model.ResponseGeneral
// @Failure      500 {object} model.ResponseGeneral
// @Router       /player/v2/admin/oculus-attestation/list [get]
func (p *playerV2HandlerImpl) ListAttestationRecordHandler(c *fiber.Ctx) (err error) {
	return fiber.NewError(fiber.StatusNotImplemented, "Attestation history is not stored by this backend")
}

type ToggleBanRequest struct {
	PlatformID               string `json:"platform_id"` // just for record, can trace with unique_id
	AttestationClaimUniqueId string `json:"unique_id" validate:"required"`
	IsBanning                bool   `json:"is_banning"` // true = ban, false = unban
	RemainingBanTime         int    `json:"remaining_time_in_minute" validate:"gte=0,lte=52560000"`
	Reason                   string `json:"reason,omitempty"`
	BanId                    string `json:"ban_id" validate:"required_if=IsBanning false"`
}

func (f *ToggleBanRequest) Validate() (errMsgs []string) {
	if err := validate.Struct(f); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

// PlayerOculusAttestationToggleBanHandler godoc
// @Summary      Toggle Player Ban Status
// @Description  Bans or Unbans a player. Pass is_banned=true to ban, false to unban.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        payload body      ToggleBanRequest true "Ban/Unban JSON"
// @Success      200 {object} model.ResponseGeneral
// @Failure      400 {object} model.ResponseGeneral
// @Failure      500 {object} model.ResponseGeneral
// @Router       /player/v2/admin/oculus-attestation/toggle-ban [post]
func (p *playerV2HandlerImpl) PlayerOculusAttestationToggleBanHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	//Parse POST JSON:
	usrInput := new(ToggleBanRequest)
	if err = c.BodyParser(usrInput); err != nil {
		resp = model.Default400Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}

	//Validate input, get error msgs according
	if errMsgs := usrInput.Validate(); len(errMsgs) > 0 {
		resp = model.Default400Resp.Copy()
		resp.Error = "Validation Error"
		resp.Messages = errMsgs
		return c.Status(resp.Status).JSON(resp)
	}

	//ban/unban with remote oculus API
	br := oculus.DeviceBanRequest{
		DeviceBan: model.DeviceBan{
			IsBanned:         usrInput.IsBanning,
			RemainingBanTime: usrInput.RemainingBanTime,
		},
		UniqueId: usrInput.AttestationClaimUniqueId,
		BanId:    usrInput.BanId,
	}

	var bResp oculus.BanResponse
	if bResp, err = p.oculusSvc.RequestOculusAttestationBan(c.Context(), &br); err != nil {
		resp = model.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	//log.Printf("ban response = %v", bResp)

	// Describe the remote operation; no history is persisted.
	esDevBanRecord := model.ESDeviceBanRecord{
		PlatformID:       usrInput.PlatformID, // just for easy backtracking
		BanID:            bResp.BanID,
		UniqueId:         usrInput.AttestationClaimUniqueId,
		RemainingBanTime: usrInput.RemainingBanTime,
		Reason:           usrInput.Reason,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
	}

	if usrInput.IsBanning {
		esDevBanRecord.Action = model.ActionBan
	} else {
		esDevBanRecord.Action = model.ActionUnban
	}

	resp = model.Default200Resp.Copy()
	resp.Data = fiber.Map{
		"ban_id":             esDevBanRecord.BanID,
		"action":             esDevBanRecord.Action,
		"remaining_ban_time": esDevBanRecord.RemainingBanTime,
		"reason":             esDevBanRecord.Reason,
		"timestamp":          esDevBanRecord.Timestamp,
		"unique_id":          esDevBanRecord.UniqueId,
	}
	return c.Status(resp.Status).JSON(resp)
}

// ListAttestationBanRecordHandler godoc
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200 {object} model.ResponseGeneral
// @Failure      400 {object} model.ResponseGeneral
// @Failure      500 {object} model.ResponseGeneral
// @Router       /player/v2/admin/oculus-attestation/ban-list [get]
func (p *playerV2HandlerImpl) ListAttestationDeviceBanRecordHandler(c *fiber.Ctx) (err error) {
	return fiber.NewError(fiber.StatusNotImplemented, "Attestation history is not stored by this backend")
}

// CheckRemoteAttestationBanStatus godoc
// @Summary Check Remote Attestation Ban Status
// @Description Check remote attestation ban status
// @Tags admin
// @Accept json
// @Produce json
// @Param unique_id query string true "Unique ID"
// @Success 200 {object} model.ResponseGeneral
// @Failure 400 {object} model.ResponseGeneral
// @Failure 500 {object} model.ResponseGeneral
// @Router /player/v2/admin/oculus-attestation/remote-ban-status/:unique_id [get]
func (p *playerV2HandlerImpl) CheckRemoteAttestationBanStatus(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	uniqueID := c.Params("unique_id", "")
	if len(uniqueID) == 0 {
		resp = model.Default400Resp.Copy()
		resp.AddErrorAsMessages(myERR.ErrMailboxInvalidPlayer)
		return c.Status(resp.Status).JSON(resp)
	}

	req := oculus.BanStatusRequest{
		UniqueId: uniqueID,
	}

	var statusResp oculus.BanStatusResponse
	if statusResp, err = p.oculusSvc.RequestOculusAttestationBanStatus(&req); err != nil {
		resp = model.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	resp = model.Default200Resp.Copy()
	if len(statusResp.Data) > 0 {
		first := statusResp.Data[0]
		log.Printf("first = %+v\n", first)
		resp.Data = fiber.Map{
			"is_banned":          first.IsBanned,
			"remaining_ban_time": first.RemainingBanTime,
			"ban_creation_date":  first.BanCreationDate,
		}
	}
	return c.Status(resp.Status).JSON(resp)
}

//	PlayerAdminListHandler godoc
//
// @Summary      Get Admin Player List
// @Description  Retrieve a paginated list of players with administrative privileges.
// @Tags player, admin
// @Produce json
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /player/v2/admin/list [get]
func (p *playerV2HandlerImpl) PlayerAdminListHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	var adms []model.Player
	if adms, err = p.playerRepo.ListAdminUser(); err != nil {
		resp = model.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	resp = model.Default200Resp.Copy()
	resp.Data = adms
	return c.Status(resp.Status).JSON(resp)
}

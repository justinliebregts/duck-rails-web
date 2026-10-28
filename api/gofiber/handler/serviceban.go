package handler

import (
	"context"
	"fmt"
	"gofiber/config"
	"gofiber/middleware"
	"gofiber/model"
	"gofiber/repository"
	"gofiber/service/cloudflare"
	"log"
	"math"
	"net/url"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// ServiceBanListActiveByPlayerHandler godoc
// @Summary Get Active ban records for me
// @Description Get Active ban records for specific player
// @Tags ban
// @Accept */*
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Success 200 {object} model.ResponseGeneral{status=int,data=model.BanInfo}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /serviceban/get [get]
func ServiceBanListActiveByPlayerHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	ses := middleware.RetrieveSession(c)

	var playerFromDB model.Player
	if playerFromDB, err = model.TryGetPlayerFromDB(&ses.SessionBase); err != nil {
		resp = model.Default400Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}

	var banInfo model.BanInfo
	if banInfo, err = model.GetPlayerServiceBanInfo(ses.ComposedID(), playerFromDB.DeviceUUID); err != nil {
		resp = model.Default400Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}

	resp = model.Default200Resp.Copy()
	resp.Data = banInfo

	if banInfo.MicBanCount > 0 || banInfo.GameBanCount > 0 {
		resp.Message = "Player has a ban."
	} else {
		resp.Message = "Player is not banned."
	}

	return c.JSON(resp)
}

// ServiceBanListAllAdminHandler godoc
// @Summary Get ban records
// @Description Get ban records for All Players
// @Tags ban
// @Accept */*
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param active_only query boolean false "Filter to only include active records"
// @Param ban_type query string false "Filter by ban type (game, mic, automated)"
// @Success 200 {object} model.ResponseGeneral{status=int,data=[]model.ServiceBanAdmin}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /serviceban/admin/all [get]
func ServiceBanListAllAdminHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	activeOnly := c.QueryBool("active_only", false)
	banType := model.BanType(c.Query("ban_type"))

	//try fetch query
	var q model.ListQuery
	if err = c.QueryParser(&q); err != nil {
		resp = model.Default400Resp.Copy()
		resp.AddMessages(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	q.Generalize()

	now := model.CurrentTimestamp()
	cnt, err := model.CountServiceBansAdminWithQuery(now, &q, activeOnly, banType)
	if err != nil {
		resp = model.Default500Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}

	var sbAdms []model.ServiceBanAdmin
	if sbAdms, err = model.GetAllServiceBansAdminWithQuery(now, &q, activeOnly, banType); err != nil {
		resp = model.Default500Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}

	for i := range sbAdms {
		sbAdms[i].FormReasonString()
		_, sbAdms[i].TimeRemaining = model.GetTimeDifferenceInHumanReadableFormat(now, sbAdms[i].EndDatetime)
	}

	resp = model.Default200Resp.Copy()
	resp.Data = sbAdms

	mt := model.PaginationMeta{
		ListQuery:  q,
		TotalCount: cnt,
	}
	mt.Generalize()

	resp.Meta = &mt
	return c.JSON(resp)
}

// ServiceBanGetPlayerAdminHandler godoc
// @Summary Get ban records
// @Description Get serviceban records for specific player
// @Tags ban
// @Accept */*
// @Produce json
// @Security ApiKeyNonStandardAuth
// @param gr_uuid path string true "GR UUID"
// @Param active_only query boolean false "Filter to only include active records"
// @Success 200 {object} model.ResponseGeneral{status=int,data=[]model.ServiceBanAdmin}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /serviceban/admin/get/{gr_uuid} [get]
func ServiceBanGetPlayerAdminHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	var uuid string
	if uuid, err = url.QueryUnescape(c.Params("gr_uuid")); err != nil {
		resp = model.Default500Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}

	nonExpiredOnly := c.QueryBool("active_only", false)

	ses := model.FormSessionBase(uuid)

	var banAdms []model.ServiceBan
	if banAdms, err = model.GetSpecificServiceBanInfoBySession(ses, nonExpiredOnly); err != nil {
		resp = model.Default500Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}

	for idx := range banAdms {
		banAdms[idx].FormReasonString()
	}

	resp = model.Default200Resp.Copy()
	resp.Data = model.NewServiceBanWithExpirationMarkSlice(banAdms)
	return c.JSON(resp)
}

// ServiceBanCreationForm is used to store UserInput
type ServiceBanCreationForm struct {
	TargetType model.BanTargetType `form:"target_type" json:"target_type" validate:"oneof=device platform"`
	TargetID   string              `form:"target_id" json:"target_id" validate:"required"`
	BanType    model.BanType       `form:"ban_type" json:"ban_type" validate:"oneof=mic automated game" default:"mic"`
	Duration   int                 `form:"duration" json:"duration" validate:"required,numeric,min=0"`
	Reason     string              `form:"reason" json:"reason"`
}

func (p *ServiceBanCreationForm) Validate() (errMsgs []string) {
	if err := validate.Struct(p); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

// ServiceBanPlayerAdminHandler godoc
// @Summary Create a service_ban record
// @Description Create a service_ban record for a Specific Player
// @Tags ban
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param form formData ServiceBanCreationForm true "Service Ban Creation Form"
// @Security ApiKeyNonStandardAuth
// @Success 200 {object} model.ResponseGeneral{status=int,data=[]model.ServiceBanAdmin}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /serviceban/admin/create [post]
func ServiceBanPlayerAdminHandler(c *fiber.Ctx) (err error) {

	var resp model.ResponseGeneral

	ses := middleware.RetrieveSession(c)

	usrInput := new(ServiceBanCreationForm)
	if err = c.BodyParser(usrInput); err != nil {
		resp = model.Default503Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	// valid input, get error msgs according
	if errMsgs := usrInput.Validate(); len(errMsgs) > 0 {
		resp = model.Default400Resp.Copy()
		resp.Messages = errMsgs
		resp.Error = "Validation Error"
		return c.Status(resp.Status).JSON(resp)
	}

	isAutomatic := usrInput.BanType == model.BanTypeAutomated
	msg := "days"
	if isAutomatic {
		msg = "minutes"
	}

	rea := map[string]string{"reason": usrInput.Reason, "By": "ServiceBanPlayerAdminHandler:" + ses.ComposedID(), "ver": config.RetrieveApiVersion()}

	if _, err = model.BanServiceAdmin(usrInput.TargetType, usrInput.TargetID, usrInput.Duration, usrInput.BanType, isAutomatic, rea); err != nil {
		resp = model.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	resp = model.Default201Resp.Copy()
	resp.AddMessages(fmt.Sprintf("ServiceBan Created for %d %s", usrInput.Duration, msg))

	return c.JSON(resp)
}

// EndServiceBanAdminHandler godoc
// @Summary End a service_ban record
// @Description End a service_ban record for a Specific BanRecord ID
// @Tags ban
// @Accept */*
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param ban_id path integer true "Ban ID"
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /serviceban/admin/end/{serban_id} [post]
func EndServiceBanAdminHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral
	log.Print("1")
	var banID uint64
	if banID, err = strconv.ParseUint(c.Params("serban_id"), 10, 64); err != nil {
		resp = model.Default503Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}
	log.Print("2")
	var rowAffected int64
	if rowAffected, err = model.EndBanServiceAdmin(banID); err != nil {
		resp = model.Default500Resp.Copy()
		resp.AddErrorAsMessages(err)
		resp.Message = "Failed to end service ban"
		return c.Status(fiber.StatusInternalServerError).JSON(resp)
	}

	resp = model.Default200Resp.Copy()
	if rowAffected > 0 {
		resp.Message = "ServiceBanID:" + strconv.FormatUint(banID, 10) + " Ended!"
	} else {
		resp.Message = "Success with no Changes"
	}

	return c.JSON(resp)
}

type IpBlacklistDeleteRequest struct {
	IP string `json:"ip" validate:"required,ip"`
}

func (f *IpBlacklistDeleteRequest) Validate() (errMsgs []string) {
	if err := validate.Struct(f); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

// ServiceBanIPBlacklistAdminDeleteHandler godoc
// @Summary Delete IP Blacklist record
// @Description Delete IP Blacklist records
// @Tags ban, admin
// @Accept json
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param request body IpBlacklistDeleteRequest true "IP Blacklist Delete Request"
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /serviceban/admin/ip-blacklist [delete]
func ServiceBanIPBlacklistAdminDeleteHandler(ipcRepo repository.IPCheckRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var resp model.ResponseGeneral

		usrInput := new(IpBlacklistDeleteRequest)
		if err := c.BodyParser(usrInput); err != nil {
			resp = model.Default400Resp.Copy()
			resp.SetErrorAndMessage(err.Error())
			return c.Status(resp.Status).JSON(resp)
		}

		// valid input, get error msgs according
		if errMsgs := usrInput.Validate(); len(errMsgs) > 0 {
			resp = model.Default400Resp.Copy()
			resp.Messages = errMsgs
			resp.Error = "Validation Error"
			return c.Status(resp.Status).JSON(resp)
		}

		// Pass Fiber's request context to the repository for timeout/cancellation support
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()

		if err := ipcRepo.UnbanIP(ctx, usrInput.IP); err != nil {
			resp = model.Default500Resp.Copy()
			resp.SetErrorAndMessage(err.Error())
			return c.Status(resp.Status).JSON(resp)
		}

		resp = model.Default200Resp.Copy()
		resp.Data = fiber.Map{
			"ts": time.Now().Unix(),
		}
		return c.JSON(resp)
	}
}

type IpBlacklistRecoverBanRequest struct {
	IpBlacklistDeleteRequest `validate:"structonly"`
}

func (f *IpBlacklistRecoverBanRequest) Validate() (errMsgs []string) {
	if err := validate.Struct(f); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

// ServiceBanIPBlacklistAdminRecoverBanHandler godoc
// @Summary Recover IP Blacklist record
// @Description Recover a soft Deleted IP Blacklist record
// @Tags ban, admin
// @Accept json
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param request body IpBlacklistRecoverBanRequest true "IP Blacklist Recover Ban Request"
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /serviceban/admin/ip-blacklist [put]
func ServiceBanIPBlacklistAdminRecoverBanHandler(ipcRepo repository.IPCheckRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var resp model.ResponseGeneral

		usrInput := new(IpBlacklistRecoverBanRequest)
		if err := c.BodyParser(usrInput); err != nil {
			resp = model.Default400Resp.Copy()
			resp.SetErrorAndMessage(err.Error())
			return c.Status(resp.Status).JSON(resp)
		}

		// valid input, get error msgs according
		if errMsgs := usrInput.Validate(); len(errMsgs) > 0 {
			resp = model.Default400Resp.Copy()
			resp.Messages = errMsgs
			resp.Error = "Validation Error"
			return c.Status(resp.Status).JSON(resp)
		}

		// Pass Fiber's request context to the repository for timeout/cancellation support
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()

		if err := ipcRepo.RecoverIPBan(ctx, usrInput.IP); err != nil {
			resp = model.Default500Resp.Copy()
			resp.SetErrorAndMessage(err.Error())
			return c.Status(resp.Status).JSON(resp)
		}

		resp = model.Default200Resp.Copy()
		resp.Data = fiber.Map{
			"ts": time.Now().Unix(),
		}
		return c.Status(resp.Status).JSON(resp)
	}
}

type IpBlacklistCreateRequest struct {
	IpBlacklistDeleteRequest `validate:"structonly"` // Ensures embedded fields are validated
	Reason                   string                  `json:"reason" validate:"required"`
}

func (f *IpBlacklistCreateRequest) Validate() (errMsgs []string) {
	if err := validate.Struct(f); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

// ServiceBanIPBlacklistAdminCreateHandler godoc
// @Summary Create IP Blacklist records
// @Description Create IP Blacklist records
// @Tags ban, admin
// @Accept json
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param request body IpBlacklistCreateRequest true "IP Blacklist Create Request"
// @Success 201 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /serviceban/admin/ip-blacklist [post]
func ServiceBanIPBlacklistAdminCreateHandler(ipcRepo repository.IPCheckRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var resp model.ResponseGeneral

		usrInput := new(IpBlacklistCreateRequest)
		if err := c.BodyParser(usrInput); err != nil {
			resp = model.Default400Resp.Copy()
			resp.SetErrorAndMessage(err.Error())
			return c.Status(resp.Status).JSON(resp)
		}

		// valid input, get error msgs according
		if errMsgs := usrInput.Validate(); len(errMsgs) > 0 {
			resp = model.Default400Resp.Copy()
			resp.Messages = errMsgs
			resp.Error = "Validation Error"
			return c.Status(resp.Status).JSON(resp)
		}

		// Pass Fiber's request context to the repository for timeout/cancellation support
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()

		if err := ipcRepo.BanIP(ctx, usrInput.IP, usrInput.Reason); err != nil {
			resp = model.Default500Resp.Copy()
			resp.SetErrorAndMessage(err.Error())
			return c.Status(resp.Status).JSON(resp)
		}

		resp = model.Default201Resp.Copy()
		resp.Data = fiber.Map{
			"ts": time.Now().Unix(),
		}
		return c.Status(resp.Status).JSON(resp)
	}
}

// ServiceBanIPBlacklistAdminHandler godoc
// @Summary Get IP Blacklist records
// @Description Get IP Blacklist records
// @Tags ban
// @Accept */*
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /serviceban/admin/ip-blacklist [get]
func ServiceBanIPBlacklistAdminHandler(ipcRepo repository.IPCheckRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var resp model.ResponseGeneral
		// Pass Fiber's request context to the repository for timeout/cancellation support
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()

		blacklist, err := ipcRepo.List(ctx)
		if err != nil {
			resp = model.Default500Resp.Copy()
			resp.SetErrorAndMessage(err.Error())
			return c.Status(resp.Status).JSON(resp)
		}

		resp = model.Default200Resp.Copy()
		resp.Data = fiber.Map{
			"blacklist": blacklist,
			"count":     len(blacklist),
			"ts":        time.Now().Unix(),
		}
		return c.JSON(resp)
	}
}

// ServiceBanCloudflareIPBlacklistAdminHandler godoc
// @Summary Get Cloudflare IP Blacklist records
// @Description Get Cloudflare IP Blacklist records
// @Tags ban, admin, cloudflare
// @Accept */*
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /serviceban/admin/cloudflare-ip-blacklist [get]
func ServiceBanCloudflareIPBlacklistAdminHandler(cipSrv cloudflare.CloudflareIPBanService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var resp model.ResponseGeneral

		var q model.ListQuery
		if err := c.QueryParser(&q); err != nil {
			resp = model.Default400Resp.Copy()
			resp.SetErrorAndMessage(err.Error())
			return c.Status(resp.Status).JSON(resp)
		}
		q.Generalize()

		// Pass Fiber's request context to the repository for timeout/cancellation support
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()

		blacklist, meta, err := cipSrv.ListBannedIPs(ctx, &q)
		if err != nil {
			resp = model.Default500Resp.Copy()
			resp.SetErrorAndMessage(err.Error())
			return c.Status(resp.Status).JSON(resp)
		}

		// Define the clean response shape
		type BannedIPDTO struct {
			ID        string `json:"id"`
			Target    string `json:"target"`
			IP        string `json:"ip"`
			Comment   string `json:"comment"`
			Mode      string `json:"mode"`
			CreatedOn string `json:"created_on"`
		}

		var cleanList []BannedIPDTO
		for _, item := range blacklist {
			cleanList = append(cleanList, BannedIPDTO{
				ID:        item.ID,
				Target:    item.Configuration.Target,
				IP:        item.IP(),
				Comment:   item.Comment(),
				Mode:      item.Mode,
				CreatedOn: item.CreatedOn,
			})
		}

		totalPage := 0
		if meta.PerPage > 0 {
			totalPage = int(math.Ceil(float64(meta.TotalCount) / float64(meta.PerPage)))
		}

		resp = model.Default200Resp.Copy()
		resp.Data = fiber.Map{
			"blacklist": cleanList,
			"count":     len(blacklist),
			"ts":        time.Now().Unix(),
			"meta": model.PaginationMeta{
				TotalCount: int64(meta.TotalCount),
				TotalPages: totalPage,
				ListQuery: model.ListQuery{
					PagePerCount: meta.PerPage,
					Page:         meta.Page,
				},
			},
		}
		return c.JSON(resp)
	}
}

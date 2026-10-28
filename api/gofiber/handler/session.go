package handler

import (
	"gofiber/config"
	myERR "gofiber/errors"
	"gofiber/model"
	myREPO "gofiber/repository"
	mySRV "gofiber/service"
	"log"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Session interface {
	SessionCheckHandler(c *fiber.Ctx) error
	SessionCreateHandler(c *fiber.Ctx) error
	SessionRefreshHandler(c *fiber.Ctx) error
	SessionActiveListHandler(c *fiber.Ctx) error
	SessionPlatformStatsHandler(c *fiber.Ctx) error

	AdminGetAllTokens(c *fiber.Ctx) error
	AdminRemoveAllTokens(c *fiber.Ctx) error
}

type sessionHandlerImpl struct {
	SesRepo myREPO.Session

	SesSrv mySRV.Session
}

func NewSessionHandler(SesRepo myREPO.Session, SesSrv mySRV.Session) Session {
	if SesRepo == nil {
		log.Fatal("SesRepo cannot be nil")
	}
	if SesSrv == nil {
		log.Fatal("SesSrv cannot be nil")
	}
	return &sessionHandlerImpl{
		SesRepo: SesRepo,
		SesSrv:  SesSrv,
	}
}

// SessionCreateHandler godoc
// @Summary Create Session for Specific Player
// @Description Crete Session using PlayerID; Check with Identity With Platform
// @Tags session
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param session_create formData model.SessionCreateForm true "Session Create Form"
// @Success 200 {object} model.ResponseGeneral{status=int,data=model.Session}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /session/create [post]
func (s *sessionHandlerImpl) SessionCreateHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	//Parse POST Input
	usrInput := new(model.SessionCreateForm)
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

	// Create report and block if not match
	if err = mySRV.HashVersionCheckAndReport(usrInput); err != nil {
		resp = model.Default403Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	//if usrInput.IsDeviceBlocked() {
	//	resp = model.Default403Resp.Copy()
	//	resp.SetErrorAndMessage(model.Banned.Error())
	//	return c.Status(resp.Status).JSON(resp)
	//}

	sb := usrInput.ToSessionBase()
	// Check if player Banned from Network services
	var banned bool
	if banned, err = model.IsGameBanned(&sb); err != nil {
		resp = model.Default503Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	} else if banned {
		resp = model.Default403Resp.Copy()
		resp.SetErrorAndMessage(myERR.Banned.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	//check if SessionBase{PlayerID, Platform} registered in `player` table
	if _, err = model.SearchBySessionBase(sb); err != nil {
		resp = model.Default400Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		time.Sleep(model.ThreeSeconds)
		return c.Status(resp.Status).JSON(resp)
	}

	////TODO: Validate platform/platform_id/platform_token
	//if errMsgs := usrInput.ValidateWithPlatform(); len(errMsgs) > 0 {
	//	resp = model.Default400Resp.Copy()
	//	resp.Error = "Platform Validation Error"
	//	resp.Messages = errMsgs
	//	return c.Status(resp.Status).JSON(resp)
	//}

	// create/update Player_Online_Session into Cache
	var session model.Session
	if session, err = model.CreatePlayerCache(usrInput.PlayerID, usrInput.Platform); err != nil {
		resp = model.Default503Resp.Copy()
		resp.Error = "Cache Error"
		resp.Message = err.Error()
		return c.Status(resp.Status).JSON(resp)
	}

	resp = model.Default201Resp.Copy()
	resp.Data = session.FormResponse()
	resp.AddMessages("Session created successfully")
	return c.JSON(resp)
}

type SessionRefreshForm struct {
	model.SessionRefresh
}

func (p *SessionRefreshForm) Validate() (errMsgs []string) {
	if err := validate.Struct(p); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

// SessionRefreshHandler godoc
// @Summary refresh token for Auth user
// @Description refresh token for Auth user
// @Tags session
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param session_refresh formData SessionRefreshForm true "Session Refresh Form"
// @Success 200 {object} model.ResponseGeneral{status=int,data=model.Session}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /session/refresh [post]
// @Router /session/refresh/{session_id}  [post]
func (s *sessionHandlerImpl) SessionRefreshHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	// param: session_id from path // TODO: Not needed
	//sessionID := c.Params("session_id")
	//if len(sessionID) <= 0 {
	//	resp = model.Default400Resp.Copy()
	//	return c.Status(resp.Status).JSON(resp)
	//}

	//Parse POST Input
	usrInput := new(SessionRefreshForm)
	if err = c.BodyParser(usrInput); err != nil {
		resp = model.Default503Resp.Copy()
		resp.Error = err.Error()
		resp.Message = err.Error()
		return c.Status(resp.Status).JSON(resp)
	}

	// valid input, get error msgs according
	if errMsgs := usrInput.Validate(); len(errMsgs) > 0 {
		resp = model.Default400Resp.Copy()
		resp.Error = "Input Validation Error"
		resp.Messages = errMsgs
		return c.Status(resp.Status).JSON(resp)
	}

	//SessionBase from Cache with RefreshToken
	var sb model.SessionBase
	if sb, err = s.SesRepo.GetRefreshData(usrInput.RefreshToken); err != nil {
		resp = model.Default400Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		time.Sleep(model.ThreeSeconds)
		return c.Status(resp.Status).JSON(resp)
	}

	// Expire previous refresh token in Cache
	if err = s.SesRepo.DeleteRefreshToken(usrInput.RefreshToken); err != nil {
		resp = model.Default503Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	// Check if player Banned from Network services
	var banned bool
	if banned, err = model.IsGameBanned(&sb); err != nil {
		resp = model.Default503Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	} else if banned {
		resp = model.Default403Resp.Copy()
		resp.SetErrorAndMessage(myERR.Banned.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	// create/update Player_Online_Session into Cache
	var res model.Session
	if res, err = model.CreatePlayerCache(sb.PlayerID, sb.Platform); err != nil {
		resp = model.Default503Resp.Copy()
		resp.Error = "Cache Error"
		resp.Message = err.Error()
		return c.Status(resp.Status).JSON(resp)
	}

	resp = model.Default200Resp.Copy()
	resp.Data = res.FormResponse()
	resp.AddMessages("Session refreshed successfully")
	return c.JSON(resp)
}

type SessionCheckForm struct {
	PlayerID    string `query:"usr" validate:"min=3"`
	Platform    string `query:"pla" validate:"min=3"`
	AccessToken string `query:"tkn" validate:"min=3"`
	SecretPhase string `query:"secret" validate:"min=3"`
}

func (f *SessionCheckForm) Validate() (errMsgs []string) {
	if err := validate.Struct(f); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

type SessionCheckResponse struct {
	ResultCode model.PhotonResultCode
	Message    string `json:",omitempty"`
	UserId     string `json:",omitempty"`
	OriginId   string `json:",omitempty"`
	Platform   string `json:",omitempty"`
	Nickname   string `json:",omitempty"`
	AuthCookie string `json:",omitempty"`
	Data       string `json:",omitempty"`
}

// SessionCheckHandler godoc
// @Summary
// @Description
// @Tags session
// @Accept */*
// @Produce json
// @Param q query SessionCheckForm true "Session Check Form"
// @Success 200 {object} model.ResponseGeneral{status=int,data=model.Session}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /photon/authenticate [get]
// @Router /fusion/authenticate [get]
func (s *sessionHandlerImpl) SessionCheckHandler(c *fiber.Ctx) (err error) {
	var resp SessionCheckResponse

	var input SessionCheckForm
	if err = c.QueryParser(&input); err != nil {
		resp.ResultCode = model.PhotonInvalidParameters
		resp.Message = err.Error()
		return c.JSON(resp)
	}

	if errMsgs := input.Validate(); len(errMsgs) > 0 {
		resp.ResultCode = model.PhotonInvalidParameters
		resp.Message = errMsgs[0]
		return c.JSON(resp)
	}

	if !config.IsPhotonSecretPhaseMatched(input.SecretPhase) {
		resp.ResultCode = model.PhotonInvalidParameters
		resp.Message = "Secret Phase Invalid!"
		return c.JSON(resp)
	}

	var sess model.Session
	sess, err = model.RetrievePlayerCache(input.AccessToken)

	// Dev Print
	if err == nil && c.Locals("base_configuration").(config.BaseConfiguration).IsEnvDevelopment {
		log.Printf("RetrievePlayerCache on %v", sess)
	}

	if err != nil {
		resp.ResultCode = model.PhotonAuthenticationFailed
		resp.Message = err.Error()
	} else if sess.PlayerID != input.PlayerID || sess.Platform != input.Platform {
		resp.ResultCode = model.PhotonAuthenticationFailed
		resp.Message = myERR.NoRowFound.Error()
	} else {
		resp.ResultCode = model.PhotonAuthenticationSuccessful
		resp.UserId = sess.ComposedID()
		resp.OriginId = sess.PlayerID
		resp.Platform = sess.Platform
	}

	return c.JSON(resp)
}

//func recordPlatformSession(sess model.Session) (int64, error) {
//	return model.CreatePlatformSessionEntry(&sess.SessionBase)
//}

type SessionStatsQuery struct {
	Start            string `query:"start"`
	End              string `query:"end"`
	DistinctByPlayer bool   `query:"distinct"`

	startDatetime time.Time
	endDatetime   time.Time
}

func (s *SessionStatsQuery) ValidateTime() (errMsgs []string) {
	var err error
	if s.startDatetime, err = model.ParseDateString(s.Start); err != nil {
		errMsgs = append(errMsgs, err.Error())
		err = nil
	}

	if s.endDatetime, err = model.ParseDateStringLastSecond(s.End); err != nil {
		errMsgs = append(errMsgs, err.Error())
		err = nil
	}

	if s.endDatetime.Before(s.startDatetime) {
		errMsgs = append(errMsgs, strconv.ErrRange.Error())
	}

	return
}

type PlatformStatsResponse struct {
	StartDatetime           model.JsonTime        `json:"start"`
	EndDatetime             model.JsonTime        `json:"end"`
	PlatformRatio           []model.PlatformRatio `json:"ratio"`
	DeviceOnlyPlatformRatio []model.PlatformRatio `json:"device_only_ratio"`

	platformCount []model.PlatformCount
}

func (p *PlatformStatsResponse) CalculateRatio() {
	p.PlatformRatio = make([]model.PlatformRatio, len(p.platformCount))
	p.DeviceOnlyPlatformRatio = []model.PlatformRatio{}

	sum := 0
	devOnlySum := 0
	for _, v := range p.platformCount {
		if v.IsDevice() {
			devOnlySum += v.Count
		} else {
			sum += v.Count
		}
	}
	sum += devOnlySum

	for i, v := range p.platformCount {
		p.PlatformRatio[i].PlatformCount = v
		p.PlatformRatio[i].Count = float32(v.Count*100) / float32(sum)

		if v.IsDevice() {
			pk := model.PlatformRatio{
				PlatformCount: v,
				Count:         float32(v.Count*100) / float32(devOnlySum),
			}

			p.DeviceOnlyPlatformRatio = append(p.DeviceOnlyPlatformRatio, pk)
		}
	}
}

// SessionPlatformStatsHandler godoc
// @Summary
// @Description
// @Tags session
// @Accept */*
// @Param q query SessionStatsQuery true "Session Stats Form"
// @Produce json
// @Success 200 {object} model.ResponseGeneral{status=int,data=PlatformStatsResponse}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /session/report [get]
func (s *sessionHandlerImpl) SessionPlatformStatsHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	var input SessionStatsQuery
	if err = c.QueryParser(&input); err != nil {
		resp = model.Default503Resp.Copy()
		resp.Message = err.Error()
		return c.Status(resp.Status).JSON(resp)
	}

	if errMsgs := input.ValidateTime(); len(errMsgs) > 0 {
		resp = model.Default400Resp.Copy()
		resp.Error = "Input Validation Error"
		resp.Messages = errMsgs
		return c.Status(resp.Status).JSON(resp)
	}

	pss := PlatformStatsResponse{
		StartDatetime: model.JsonTime{input.startDatetime},
		EndDatetime:   model.JsonTime{input.endDatetime},
		PlatformRatio: []model.PlatformRatio{},
	}

	var pc []model.PlatformCount
	if pc, err = model.GetPlatformUsageStatsFromDateRange(pss.StartDatetime.Time, pss.EndDatetime.Time, input.DistinctByPlayer); err == nil {
		pss.platformCount = pc
		pss.CalculateRatio()
	}

	resp = model.Default200Resp.Copy()
	resp.Data = pss
	resp.AddMessages("Session Stats retrieved successfully")
	if input.DistinctByPlayer {
		resp.AddMessages("distinct By PlayerID")
	}

	return c.JSON(resp)
}

// SessionActiveListHandler godoc
// @Summary List all player sessions
// @Description
// @Tags session
// @Accept */*
// @Produce json
// @Success 200 {object} model.ResponseGeneral{status=int,data=[]string}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /session/online [get]
func (s *sessionHandlerImpl) SessionActiveListHandler(c *fiber.Ctx) (err error) {
	resp := model.Default200Resp.Copy()
	resp.Data = model.AllSessions() //resp.Data = (&myDOM.SessionBase{}).All()
	resp.AddMessages("Sessions retrieved successfully")
	return c.JSON(resp)
}

// AdminGetAllTokens godoc
// @Summary      Get all tokens for a specific user
// @Description  Get all active sessions for a user identified by platform and platform_id
// @Tags         session
// @Produce json
// @Param        platform     path      string  true  "Platform (e.g., google, apple)"
// @Param        platform_id  path      string  true  "Platform User ID"
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /session/admin/{platform}/{platform_id} [Get]
func (s *sessionHandlerImpl) AdminGetAllTokens(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	platform := c.Params("platform")
	platformID := c.Params("platform_id")

	sb := model.NewSessionBase(platform, platformID)
	if errMsgs := sb.Validate(); len(errMsgs) > 0 {
		resp = model.Default400Resp.Copy()
		resp.Error = "Target Player Validation Error"
		resp.Messages = errMsgs
		return c.Status(resp.Status).JSON(resp)
	}
	log.Printf("SessionBase:%+v", sb)

	var accessToken string
	if accessToken, err = s.SesRepo.GetAccessTokenByUUID(sb); err != nil {
		resp = model.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	log.Printf("accessToken:%v", accessToken)

	var session model.Session
	if session, err = s.SesRepo.GetSession(accessToken); err != nil {
		resp = model.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	log.Printf("session:%+v", session)

	// double check with refresh token
	var sb2 model.SessionBase
	if sb2, err = s.SesRepo.GetRefreshData(session.RefreshToken); err == nil {
		if sb2.ComposedID() != sb.ComposedID() {
			resp = model.Default500Resp.Copy()
			resp.SetErrorAndMessage("Session integrity mismatch: token belongs to a different user")
			return c.Status(resp.Status).JSON(resp)
		}
	}

	resp = model.Default200Resp.Copy()
	resp.Data = fiber.Map{
		"access_token":  accessToken,
		"refresh_token": session.RefreshToken,
		"session_base":  sb,
	}
	return c.Status(resp.Status).JSON(resp)
}

// AdminRemoveAllTokens godoc
// @Summary      Remove all tokens for a specific user
// @Description  Revokes all active sessions for a user identified by platform and platform_id
// @Tags         session
// @Produce json
// @Param        platform     path      string  true  "Platform (e.g., google, apple)"
// @Param        platform_id  path      string  true  "Platform User ID"
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /session/admin/{platform}/{platform_id} [Delete]
func (s *sessionHandlerImpl) AdminRemoveAllTokens(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	platform := c.Params("platform")
	platformID := c.Params("platform_id")

	sb := model.NewSessionBase(platform, platformID)
	if errMsgs := sb.Validate(); len(errMsgs) > 0 {
		resp = model.Default400Resp.Copy()
		resp.Error = "Target Player Validation Error"
		resp.Messages = errMsgs
		return c.Status(resp.Status).JSON(resp)
	}

	if err = s.SesSrv.RevokeFullSession(sb); err != nil {
		resp = model.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	resp = model.Default200Resp.Copy()
	resp.Message = "All tokens removed successfully"
	return c.Status(resp.Status).JSON(resp)
}

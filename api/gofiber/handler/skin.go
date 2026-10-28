package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gofiber/model"
	"strconv"
)

// SkinAdminGetAll godoc
// @Summary
// @Description
// @Tags skin
// @Accept */*
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /skin/admin/all [get]
// @Router /skin/admin/ [get]
func SkinAdminGetAll(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	var ss []model.SkinStyle
	if ss, err = model.GetAllSkinStyles(); err != nil {
		resp = model.Default404Resp.Copy()
		resp.Error = err.Error()
		resp.Messages = []string{"No skins found"}
	} else {
		resp = model.Default200Resp.Copy()
		resp.Data = ss
		resp.Messages = []string{"Skins found"}
	}

	return c.Status(resp.Status).JSON(resp)
}

type SkinCreateForm struct {
	ID   uint   `form:"id" json:"id"`
	Name string `form:"name" json:"name" validate:"required,min=3"`
}

func (s *SkinCreateForm) Validate() (errMsgs []string) {
	if err := validate.Struct(s); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

// SkinAdminCreate godoc
// @Summary
// @Description
// @Tags skin
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param form formData SkinCreateForm true "Skin Creation Form"
// @Security ApiKeyNonStandardAuth
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /skin/admin/create [post]
func SkinAdminCreate(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	//Parse POST Input
	usrInput := new(SkinCreateForm)
	if err = c.BodyParser(usrInput); err != nil {
		resp = model.Default503Resp.Copy()
		resp.Error = err.Error()
		resp.Message = err.Error()
		return c.Status(resp.Status).JSON(resp)
	}

	// valid input, get error msgs according
	if errMsgs := usrInput.Validate(); len(errMsgs) > 0 {
		resp = model.Default400Resp.Copy()
		resp.Error = "Validation Error"
		resp.Messages = errMsgs
		return c.Status(resp.Status).JSON(resp)
	}

	var ss model.SkinStyle
	if ss, err = model.CreateSkinStyle(usrInput.ID, usrInput.Name); err != nil {
		resp = model.Default404Resp.Copy()
		resp.Error = err.Error()
		resp.Messages = []string{"Unable to create skin."}
	} else {
		resp = model.Default200Resp.Copy()
		resp.Data = ss.ID
		resp.Messages = []string{"Skins created"}
	}

	return c.Status(resp.Status).JSON(resp)
}

type SkinUpdateForm struct {
	SkinCreateForm
}

func (s *SkinUpdateForm) Validate() (errMsgs []string) {
	if err := validate.Struct(s); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

// SkinAdminUpdate godoc
// @Summary
// @Description
// @Tags skin
// @Accept application/x-www-form-urlencoded
// @Produce json
// @param skin_id path int true "SkinStyle ID"
// @Param form formData SkinUpdateForm true "Skin Update Form"
// @Security ApiKeyNonStandardAuth
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /skin/admin/update/{skin_id} [post]
func SkinAdminUpdate(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	var sid uint64
	if sid, err = strconv.ParseUint(c.Params("skin_id", "-1"), 10, 32); err != nil {
		resp = model.Default404Resp.Copy()
		resp.Error = err.Error()
		return
	}

	//Parse POST Input
	usrInput := new(SkinUpdateForm)
	if err = c.BodyParser(usrInput); err != nil {
		resp = model.Default503Resp.Copy()
		resp.Error = err.Error()
		resp.Message = err.Error()
		return c.Status(resp.Status).JSON(resp)
	}

	// valid input, get error msgs according
	if errMsgs := usrInput.Validate(); len(errMsgs) > 0 {
		resp = model.Default400Resp.Copy()
		resp.Error = "Validation Error"
		resp.Messages = errMsgs
		return c.Status(resp.Status).JSON(resp)
	}

	if err = model.UpdateSkinStyle(uint(sid), usrInput.Name); err != nil {
		resp = model.Default404Resp.Copy()
		resp.Error = err.Error()
		resp.Messages = []string{"Unable to update skin."}
	} else {
		resp = model.Default200Resp.Copy()
		resp.Data = sid
		resp.Messages = []string{"Skin updated successfully"}
	}

	return c.Status(resp.Status).JSON(resp)
}

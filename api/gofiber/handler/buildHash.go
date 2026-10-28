package handler

import (
	myERR "gofiber/errors"
	"gofiber/model"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type CreateBuildHashForm struct {
	Version string   `form:"version" validate:"required" example:"0.0.1"`
	Hash    string   `form:"hash" validate:"required" example:"abcdefg"`
	ObbHash []string `form:"obb_hash" example:"abcd,efgh"`
}

func (c *CreateBuildHashForm) ToCore() model.BuildHashCore {
	res := model.BuildHashCore{Version: c.Version, Hash: c.Hash}
	res.FillObbHash(c.ObbHash)
	return res
}

func (c *CreateBuildHashForm) Validate() (errMsgs []string) {
	if err := validate.Struct(c); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

// BuildHashAdminCreateHandler godoc
// @Summary
// @Description
// @Tags buildhash
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param form formData CreateBuildHashForm false "Create BuildHash Form"
// @Security ApiKeyNonStandardAuth
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /buildhash/admin/create [post]
func BuildHashAdminCreateHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	usrInput := new(CreateBuildHashForm)
	if err = c.BodyParser(usrInput); err != nil {
		resp = model.Default503Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	// valid input, get error msgs according
	if errMsgs := usrInput.Validate(); len(errMsgs) > 0 {
		resp = model.Default400Resp.Copy()
		resp.Messages = errMsgs
		resp.Error = myERR.Invalidated.Error()
		return c.Status(resp.Status).JSON(resp)
	}

	if err = model.CreateHashValid(usrInput.ToCore()); err != nil {
		resp = model.Default500Resp.Copy()
		resp.Error = err.Error()
		return c.Status(resp.Status).JSON(resp)
	}

	resp = model.Default201Resp.Copy()
	resp.AddMessages("Hash saved")
	return c.Status(resp.Status).JSON(resp)
}

// BuildHashValidateHandler godoc
// @Summary
// @Description
// @Tags buildhash
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param form formData model.ValidateBuildHashForm false "Validate BuildHash Form"
// @Security ApiKeyNonStandardAuth
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Failure 503 {object} model.ResponseGeneral{status=int}
// @Router /buildhash/validate [post]
func BuildHashValidateHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	usrInput := new(model.ValidateBuildHashForm)
	if err = c.BodyParser(usrInput); err != nil {
		resp = model.Default503Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	// valid input, get error msgs according
	if errMsgs := usrInput.Validate(); len(errMsgs) > 0 {
		resp = model.Default400Resp.Copy()
		resp.Messages = errMsgs
		resp.Error = myERR.Invalidated.Error()
		return c.Status(resp.Status).JSON(resp)
	}

	bhc := model.NewBuildHashCore(usrInput)
	hashVersionMatched, obbhashHashVersionMatched := model.IsHashObbHashVersionMatch(bhc)
	resp = model.Default200Resp.Copy()
	resp.Data = map[string]interface{}{
		"hash_version_matched":          hashVersionMatched,
		"hash_version_obb_hash_matched": obbhashHashVersionMatched,
	}
	return c.JSON(resp)
}

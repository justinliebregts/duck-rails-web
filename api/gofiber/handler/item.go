package handler

import (
	"database/sql"
	myERR "gofiber/errors"
	"gofiber/model"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ItemFormBase struct {
	Name           string `form:"name" json:"name" validate:"required"`
	AssetID        string `form:"asset_id" json:"asset_id" validate:"required"`
	WeaponType     string `form:"weapon_type" json:"weapon_type"`
	SkinStyleID    int    `form:"skin_style_id" json:"skin_style_id" validate:"required"`
	IsDefault      int    `form:"is_default" json:"is_default"`
	DevOwned       int    `form:"dev_owned" json:"dev_owned"`
	CanOwnMultiple int    `form:"can_own_multiple" json:"can_own_multiple"`
	GRReady        int    `form:"gr_ready" json:"gr_ready"`
}

func (p *ItemFormBase) Validate() (errMsgs []string) {
	if err := validate.Struct(p); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

func (p *ItemFormBase) apply(itm *model.Item) {
	itm.Name = p.Name
	itm.SkinStyleID = p.SkinStyleID
	itm.IsDefault = p.IsDefault
	itm.DevOwned = p.DevOwned
	itm.CanOwnMultiple = p.CanOwnMultiple
	itm.GRReady = p.GRReady
	itm.AssetID = sql.NullString{String: p.AssetID, Valid: true}
	if val, err := strconv.Atoi(p.WeaponType); err == nil && val > 0 {
		itm.WeaponType = sql.NullInt16{Int16: int16(val), Valid: true}
	}
}

type ItemCreationForm struct {
	ItemFormBase
	Type int `form:"type" json:"type" validate:"required"`
}

func (p *ItemCreationForm) ToItem() (model.Item, error) {
	if p.Type != model.ItemTypeWeapon && p.Type != model.ItemTypeAvatar {
		return model.Item{}, myERR.ErrItemTypeNotAllowed
	}
	itm := model.Item{Type: p.Type}
	p.apply(&itm)
	if p.Type == model.ItemTypeAvatar {
		itm.WeaponType = sql.NullInt16{}
	}
	return itm, nil
}

// ItemGetAllHandler godoc
// @Summary List avatar and weapon items
// @Tags item
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /item/admin/all [get]
func ItemGetAllHandler(c *fiber.Ctx) error {
	var resp model.ResponseGeneral
	rows, err := model.AllItems()
	if err != nil {
		resp = model.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	resp = model.Default200Resp.Copy()
	resp.Data = rows
	resp.AddMessages("Items found")
	return c.Status(resp.Status).JSON(resp)
}

// ItemListWeaponsHandler godoc
// @Summary List weapon types
// @Tags item
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /item/admin/weapons [get]
func ItemListWeaponsHandler(c *fiber.Ctx) error {
	var resp model.ResponseGeneral
	rows, err := model.ListWeapons()
	if err != nil {
		resp = model.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	resp = model.Default200Resp.Copy()
	resp.Data = rows
	return c.Status(resp.Status).JSON(resp)
}

// ItemPostCreateHandler godoc
// @Summary Create an avatar or weapon item
// @Tags item
// @Accept json
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param form body ItemCreationForm true "item"
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Router /item/admin/create [post]
func ItemPostCreateHandler(c *fiber.Ctx) error {
	var resp model.ResponseGeneral
	var form ItemCreationForm
	if err := c.BodyParser(&form); err != nil {
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
	item, err := form.ToItem()
	if err != nil {
		resp = model.Default400Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	created, err := model.CreateItem(item)
	if err != nil {
		resp = model.Default400Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	resp = model.Default200Resp.Copy()
	resp.Data = created.ID
	resp.AddMessages("Item created successfully")
	return c.Status(resp.Status).JSON(resp)
}

// ItemPostUpdateHandler godoc
// @Summary Update an avatar or weapon item
// @Tags item
// @Accept json
// @Produce json
// @Security ApiKeyNonStandardAuth
// @Param item_id path int true "item id"
// @Param form body ItemFormBase true "item"
// @Success 200 {object} model.ResponseGeneral{status=int}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 404 {object} model.ResponseGeneral{status=int}
// @Router /item/admin/update/{item_id} [post]
func ItemPostUpdateHandler(c *fiber.Ctx) error {
	var resp model.ResponseGeneral
	itemID, err := strconv.ParseUint(c.Params("item_id", ""), 10, 64)
	if err != nil {
		resp = model.Default400Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	existing, err := model.GetItem(uint(itemID))
	if err != nil {
		resp = model.Default404Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	if !existing.IsSkin() {
		resp = model.Default400Resp.Copy()
		resp.SetErrorAndMessage(myERR.ErrItemTypeNotAllowed.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	var form ItemFormBase
	if err = c.BodyParser(&form); err != nil {
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
	form.apply(&existing)
	if existing.Type == model.ItemTypeAvatar {
		existing.WeaponType = sql.NullInt16{}
	}
	if err = existing.Update(); err != nil {
		resp = model.Default400Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	resp = model.Default200Resp.Copy()
	resp.AddMessages("Item updated successfully")
	return c.Status(resp.Status).JSON(resp)
}

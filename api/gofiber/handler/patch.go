package handler

import (
	"gofiber/model"
	myREPO "gofiber/repository"

	"github.com/gofiber/fiber/v2"
)

type PatchHandler interface {
	PatchAccountRenameHandler(c *fiber.Ctx) error
}

type PatchHandlerImpl struct {
	dbPatchRepo myREPO.DBPatchRepository
}

func NewPatchHandler(dbPR myREPO.DBPatchRepository) PatchHandler {
	return &PatchHandlerImpl{
		dbPatchRepo: dbPR,
	}
}

type PatchAccountRenameForm struct {
	FromPlatform   string `form:"from_pla"`
	FromPlatformID string `form:"from_pla_id"`
	ToPlatform     string `form:"to_pla"`
	ToPlatformID   string `form:"to_pla_id"`
}

func (p *PatchAccountRenameForm) FromSessionBase() model.SessionBase {
	return model.SessionBase{
		Platform: p.FromPlatform,
		PlayerID: p.FromPlatformID,
	}
}

func (p *PatchAccountRenameForm) ToSessionBase() model.SessionBase {
	return model.SessionBase{
		Platform: p.ToPlatform,
		PlayerID: p.ToPlatformID,
	}
}

// PatchAccountRenameHandler godoc
// @Summary rename account from A to B
// @Description
// @Tags player
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param registration formData PatchAccountRenameForm true "Player Registration Form"
// @Success 200 {object} model.ResponseGeneral{status=int,data=model.Session}
// @Failure 400 {object} model.ResponseGeneral{status=int}
// @Failure 500 {object} model.ResponseGeneral{status=int}
// @Router /patch/rename [post]
func (p *PatchHandlerImpl) PatchAccountRenameHandler(c *fiber.Ctx) (err error) {
	var resp model.ResponseGeneral

	//Parse POST Input
	usrInput := new(PatchAccountRenameForm)
	if err = c.BodyParser(usrInput); err != nil {
		resp = model.Default400Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}

	//// valid input, get error msgs according
	//if errMsgs := usrInput.Validate(); len(errMsgs) > 0 {
	//	resp = model.Default400Resp.Copy()
	//	resp.Error = "Validation Error"
	//	resp.Messages = errMsgs
	//	return c.Status(resp.Status).JSON(resp)
	//}

	if err = p.dbPatchRepo.TransactionRenameUser(usrInput.FromSessionBase(), usrInput.ToSessionBase()); err != nil {
		resp = model.Default500Resp.Copy()
		resp.AddErrorAsMessages(err)
		return c.Status(resp.Status).JSON(resp)
	}

	//ES log this

	resp = model.Default200Resp.Copy()
	return c.JSON(resp)
}

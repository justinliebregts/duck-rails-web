package model

import (
	myERR "gofiber/errors"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-version"
)

var (
	clientVersionConstraintStr = ">= 0.9.1, < 1.4"
)

type ValidateBuildHashForm struct {
	Version string   `form:"version" validate:"required" example:"0.0.1"`
	Hash    string   `form:"session" validate:"required" example:"abcdefg"`
	ObbHash []string `form:"oculus_tag" example:"abcdefg,123456789"`
}

func (v *ValidateBuildHashForm) CheckClientVersion() (err error) {

	var clientVersion *version.Version
	if clientVersion, err = version.NewVersion(v.Version); err != nil {
		return
	}

	//log.Println(clientVersion.Segments())

	if len(clientVersion.Segments()) > 3 {
		err = myERR.ErrClientVersionInvalid
		return
	}

	//Version Constraints
	var constraints version.Constraints
	if constraints, err = version.NewConstraint(clientVersionConstraintStr); err != nil {
		return
	}

	//log.Println(constraints.String())

	if !constraints.Check(clientVersion) {
		err = myERR.ErrClientVersionTooOld
	}

	return
}

func (v *ValidateBuildHashForm) Validate() (errMsgs []string) {
	if err := validate.Struct(v); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

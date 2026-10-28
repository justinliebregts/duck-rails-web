package model

import (
	"os"

	"github.com/go-playground/validator/v10"
)

var (
	OculusBypassNoncePhrase = "5bf2efe6f0f269297b3347880342b97d8e989c6960c13811b0b55423"
	skippingBuildPlatform   = map[string]interface{}{"pcvr": true}
)

func LoadSessionSettingsFromEnv() {
	OculusBypassNoncePhrase = os.Getenv("OCULUS_BYPASS_NONCE")
}

type SessionCreateForm struct {
	SessionBase
	BuildPlatform     string `form:"build_platform" json:"build_platform,omitempty" example:"pcvr"`
	PlatformID        string `form:"platform_id" json:"platform_id" validate:"required" example:"fake-platform-id"`
	PlatformToken     string `form:"platform_token" json:"platform_token" validate:"required" example:"fake-platform-token"`
	OculusBypassNonce string `form:"oculus_bypass_nonce" json:"oculus_bypass_nonce"`
	DeviceUuid        string `form:"player_type" json:"device_uuid"  example:"abcdefghijk"`
	DeviceName        string `form:"device" json:"device" example:"Xiaomi M2007J20CG"`
	ValidateBuildHashForm
}

func (p *SessionCreateForm) Validate() (errMsgs []string) {
	if err := validate.Struct(p); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

func (p *SessionCreateForm) SkipForBuildPlatform() (shouldSkip bool) {
	if len(p.BuildPlatform) <= 0 {
		return
	}

	if _, exist := skippingBuildPlatform[p.BuildPlatform]; !exist {
		return
	}

	return true
}

func (p *SessionCreateForm) ToSessionBase() SessionBase {
	return SessionBase{PlayerID: p.PlayerID, Platform: p.Platform}
}

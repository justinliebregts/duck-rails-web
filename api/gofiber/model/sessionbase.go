package model

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

type SessionBase struct {
	PlayerID string `form:"player_id" json:"player_id,omitempty" validate:"required,min=8" example:"tester_player_id_001"`
	Platform string `form:"platform" json:"platform,omitempty" validate:"required,oneof=OCULUS EMAIL PICO YVR GOOGLE APPLE STEAM DEBUGGER ROOMBOT" example:"OCULUS"`
}

func NewSessionBase(platform, platformID string) SessionBase {
	return SessionBase{PlayerID: platformID, Platform: strings.ToUpper(platform)}
}

func FormSessionBase(grUUID string) SessionBase {
	pid, pla := SeparateKey(grUUID)
	return SessionBase{PlayerID: pid, Platform: pla}
}

func (sb *SessionBase) Validate() (errMsgs []string) {
	if err := validate.Struct(sb); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

func (sb *SessionBase) IsApplePlayer() bool {
	return sb.Platform == PlatformApple.Name
}

//func (sb *SessionBase) CreatePlayerCache() (ses Session, err error) {
//	ses = BuildSessionFromSessionBase(*sb)
//	return ses, ses.SaveToCache()
//}

func (sb *SessionBase) ComposedID() string {
	return CombinedKey(sb.PlayerID, sb.Platform)
}

func (sb *SessionBase) ComposedIDSafe() string {
	// Replace the "|" from CombinedKey with a "-"
	return strings.ReplaceAll(sb.ComposedID(), "|", "-")
}

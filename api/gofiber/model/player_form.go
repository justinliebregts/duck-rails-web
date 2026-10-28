package model

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var (
	invalidRegistrationPlatform = map[string]any{
		"ROOMBOT":  1,
		"DEBUGGER": 2,
	}

	// deviceNameBlacklist are devices that blocked at player/social-account or player/register phase
	deviceNameBlacklist = map[string]interface{}{
		"Xiaomi M2007J20CG": true,
		"Xiaomi SM-S928B":   true,
	}
)

// PlayerRegistrationForm is used to Retrieve and Validate UserInput
type PlayerRegistrationForm struct {
	SessionCreateForm
	Name   string `form:"name" validate:"required,min=1" example:"tester_player_name"`
	Region string `form:"region" validate:"required,min=1"  example:"-"`
	//Device string `form:"device" validate:"required"  example:"Oculus_Quest_2"`

	IPs []string `json:"-"`
	//AccountJustCreated bool     `json:"-"`
}

func (p *PlayerRegistrationForm) FillIPs(c *fiber.Ctx) {
	if c == nil {
		return
	}

	p.IPs = c.IPs()

	if len(p.IPs) == 0 {
		p.IPs = []string{c.IP()}
	}
}

//func (p *PlayerRegistrationForm) IsPlatformAccountRegistered() bool {
//	// check if p.Platform - p.PlatformID exist in DB
//	if _, err := model.RetrieveFromDB(p.ComposedID()); err == nil {
//		return true
//	}
//
//	return false
//}

//func (p *PlayerRegistrationForm) IsApplePlayer() bool {
//	return p.Platform == model.PlatformApple.Name
//}

func (p *PlayerRegistrationForm) IsDeviceNameBlocked() bool {

	// blocking this specific hacker device temporarily
	if _, ok := deviceNameBlacklist[p.DeviceName]; ok {
		//log.Printf("BLOCKING this device %+v", p.DeviceName)
		return true
	}
	//log.Printf("Not block this device %+v", p.DeviceName)
	return false
}

func (p *PlayerRegistrationForm) Validate() (errMsgs []string) {
	if err := validate.Struct(p); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}

	// NOT allow certain platform when registering
	if _, ok := invalidRegistrationPlatform[p.Platform]; ok {
		errMsgs = append(errMsgs, "Invalid Platform "+p.Platform)
	}

	return
}

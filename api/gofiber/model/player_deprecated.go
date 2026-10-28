package model

import (
	myERR "gofiber/errors"
)

func (p *Player) IsGameBanned() (bool, error) {

	//return IsPlayerGameBan(p.ID)
	if sbs, err := GetSpecificActiveServiceBanInfo(BanTypeGame, p.ID, p.DeviceUUID); err == nil {
		if len(sbs) <= 0 {
			return false, nil
		}
	} else {
		//log.Printf("%v IsGameBanned.GetSpecificServiceBanInfo() GotErr:%#v", p.ID, err)
		return true, err
	}

	return true, myERR.Unexpected
}

// IsInGameBan tells if this player has unexpired gameBans
func (p *Player) IsInGameBan() (banned bool, err error) {
	banned = true

	var cnt int64
	//if _, err = GetPlayerBan(p.ID); err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
	if cnt, err = GetTargetIDAnyGameBan(p.ID, p.DeviceUUID); err != nil {
		return
	}

	if cnt <= 0 {
		banned = false
	}

	return
}

func (p *Player) BanInfo() (bInfo BanInfo, err error) {
	return GetPlayerServiceBanInfo(p.ID, p.DeviceUUID) // return GetPlayerBanInfo(p.ID)
}

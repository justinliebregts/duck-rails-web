package model

import "strings"

var (
	PlatformOculus = PlatformCodeName{"", "OCULUS"}
	PlatformIqiyi  = PlatformCodeName{"i", "IQIYI"}
	PlatformPico   = PlatformCodeName{"p", "PICO"}
	PlatformEMail  = PlatformCodeName{"e", "EMAIL"}
	PlatformGoogle = PlatformCodeName{"g", "GOOGLE"}
	PlatformApple  = PlatformCodeName{"a", "APPLE"}
	PlatformSteam  = PlatformCodeName{"s", "STEAM"}
	//PlatformSteamVR    = PlatformCodeName{"t", "STEAM_VR"}
	PlatformYvr     = PlatformCodeName{"y", "YVR"}
	PlatformRoombot = PlatformCodeName{"z", "ROOMBOT"}
	PlatformUnknown = PlatformCodeName{"", "UnKnOwN"}
)

type PlatformCodeName struct {
	Code string
	Name string
}

func CombinedKey(playerID string, platform string) string {
	switch platform {
	case PlatformOculus.Name:
		return playerID
	case PlatformEMail.Name:
		return "e|" + playerID
	case PlatformPico.Name:
		return "p|" + playerID
	case PlatformIqiyi.Name:
		return "i|" + playerID
	case PlatformYvr.Name:
		return "y|" + playerID
	case PlatformApple.Name:
		return "a|" + playerID
	case PlatformGoogle.Name:
		return "g|" + playerID
	case PlatformSteam.Name:
		return "s|" + playerID
	default:
		return "z|" + playerID
	}

	//log.Printf("Unexpected Platform: %v", platform)
	//return playerID
}

func SeparateKey(composedPlayerUUID string) (playerID string, platform string) {
	fields := strings.Split(composedPlayerUUID, "|")
	fldCount := len(fields)

	if fldCount < 1 || fldCount > 2 {
		playerID = composedPlayerUUID
		platform = PlatformUnknown.Name
		return
	}

	// no | : OCULUS
	if fldCount == 1 {
		playerID = composedPlayerUUID
		platform = PlatformOculus.Name
		return
	}

	// exactly 2 fields
	playerID = fields[1]
	switch fields[0] {
	case PlatformEMail.Code:
		platform = PlatformEMail.Name
	case PlatformGoogle.Code:
		platform = PlatformGoogle.Name
	case PlatformApple.Code:
		platform = PlatformApple.Name
	case PlatformPico.Code:
		platform = PlatformPico.Name
	case PlatformIqiyi.Code:
		platform = PlatformIqiyi.Name
	case PlatformYvr.Code:
		platform = PlatformYvr.Name
	case PlatformSteam.Code:
		platform = PlatformSteam.Name
	case PlatformRoombot.Code:
		platform = PlatformRoombot.Name
	}
	return
}

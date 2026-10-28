package service

import (
	"errors"
	"gofiber/config"
	myDB "gofiber/database/mysql"
	myMDL "gofiber/model"
	myREPO "gofiber/repository"
	"log"

	"gorm.io/gorm"
)

type Player interface {
	SyncPlayerRecord(usrInput *myMDL.PlayerRegistrationForm) (*myMDL.Player, bool, error)
	IsDeviceBlocked(p *myMDL.SessionCreateForm) bool
}

func NewPlayerService(playerRepo myREPO.PlayerRepository, dbPatchRepo myREPO.DBPatchRepository) Player {
	return &playerServiceImpl{
		playerRepo:  playerRepo,
		dbPatchRepo: dbPatchRepo,
	}
}

type playerServiceImpl struct {
	playerRepo  myREPO.PlayerRepository
	dbPatchRepo myREPO.DBPatchRepository
}

func (s *playerServiceImpl) CreatePlatformSessionEntry(sb *myMDL.SessionBase) (rowAffected int64, err error) {

	entry := myMDL.PlatformSessionEntry{
		SessionBase: *sb,
	}

	result := myDB.DBConn.Create(&entry)

	return result.RowsAffected, result.Error
}

func (s *playerServiceImpl) SyncPlayerRecord(usrInput *myMDL.PlayerRegistrationForm) (playerPtr *myMDL.Player, isNewPlayer bool, err error) {
	isNewPlayer = false
	// Attempt to find existing player
	var localErr error
	playerPtr, localErr = s.playerRepo.RetrieveFromDB(usrInput.ComposedID())
	if localErr != nil {
		if errors.Is(localErr, gorm.ErrRecordNotFound) {
			//cannot load player from DB, consider new player
			newPlayer := myMDL.NewPlayer(usrInput)
			playerPtr = &newPlayer
			isNewPlayer = true
		} else {
			// real db error
			err = localErr
			return
		}
	} else {
		// Logic for Existing Player - Updating "Volatile" data
		playerPtr.Update(usrInput)
	}

	// update player info
	playerPtr.FillIPFromIPs()

	// Persistence
	err = s.playerRepo.SaveToDB(playerPtr)

	return
}

//
//func (s *playerServiceImpl) ValidateWithDevicePlatform(p *myDOM.SessionCreateForm) (errMsgs []string) {
//
//	if p.SkipForBuildPlatform() {
//		return
//	}
//
//	if !config.ShouldVerifyPlatformAccountWithRemoteAPI {
//		return
//	}
//
//	switch p.Platform {
//	default:
//		//log.Print("nothing")
//	case config.PlatformOculus.Name:
//		if p.OculusBypassNonce == myDOM.OculusBypassNoncePhrase {
//			return
//		}
//
//		q := oculus.UserNonceValidateQuery{}
//		q.Build(p.PlatformID, p.PlatformToken)
//
//		espr := myMDL.PlatformVerificationRecord{
//			PlatformID: p.PlatformID,
//			Platform:   config.PlatformOculus.Name,
//			Timestamp:  myMDL.CurrentTimestamp(),
//			Comment:    map[string]string{"nonce": p.PlatformToken},
//			AppSource:  myES.AppSource,
//		}
//
//		if res, err := oculus.RequestOculusUserNonceValidate(q); err == nil {
//			espr.OculusResult = res
//			if res.IsValid {
//				delete(espr.Comment, "nonce")
//			} else {
//				errMsgs = append(errMsgs, "Not a Valid Oculus User")
//			}
//		} else if errors.Is(err, myERR.NoFound) {
//			// special case !IsValid: err was nil and res having data
//			errMsgs = append(errMsgs, "Not a Valid Oculus User")
//		} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
//			espr.TryAddComment("Timeout", "ToCache")
//			espr.TryAddComment("Error", err.Error())
//
//			//// NonceCheck cannot be delayed since nonce changes all the time
//			//// considered okay if Timeout happened, tmp save info for CrobJobCheck
//			//_ = p.CreateNonceChecksToCache()
//		} else {
//			errMsgs = append(errMsgs, err.Error())
//			espr.TryAddComment("Error", err.Error())
//		}
//
//		// create ES Record without blocking
//		_, _ = espr.CreateElasticTypedDocument()
//	}
//
//	return
//}

func (s *playerServiceImpl) IsDeviceBlocked(p *myMDL.SessionCreateForm) bool {

	//// blocking this specific hacker device temporarily
	//if _, ok := deviceNameBlacklist[p.DeviceName]; ok {
	//	return true
	//}

	//// DeviceUuid is used for check Device ServiceBan
	if len(p.DeviceUuid) <= 0 {
		p.DeviceUuid = "fake_device_uuid" //return false
	}

	if count, err := myMDL.GetTargetIDAnyGameBan(p.ComposedID(), p.DeviceUuid); err == nil && count > 0 {
		return true
	}

	return false
}

func HashVersionCheckAndReport(p *myMDL.SessionCreateForm) error {

	if !config.IsPlatformBuildHashToCheck(p.Platform) {
		return nil
	}

	//if p.IsHashVersionMatch() {
	//	return nil
	//}

	bhc := myMDL.NewBuildHashCore(&p.ValidateBuildHashForm)
	verHashMatched, obbHaHaVerMatched := myMDL.IsHashObbHashVersionMatch(bhc)
	if verHashMatched {
		if obbHaHaVerMatched {
			return nil
		}

		// verHashMatched && !obbHashHashVersionMatched
		log.Printf("obb hash mismatch player=%s", p.SessionBase.ComposedID())
		return nil
	}

	// !verHashMatched && !obbHashHashVersionMatched
	// auto ban them.  Not sure this was working 100% due to a lot of people being banned
	_, lerr := myMDL.BanServiceAdminWithDuration(
		myMDL.BanTargetPlatformId,
		p.SessionBase.ComposedID(),
		myMDL.BanForHashVersion,
		myMDL.BanTypeGame,
		GetBanReasonByHash())
	if lerr != nil {
		log.Printf("HashVersionCheckAndReport BanServiceAdminWithDuration with err:%#v", lerr)
	}

	log.Printf("build hash mismatch player=%s", p.SessionBase.ComposedID())
	return nil
}

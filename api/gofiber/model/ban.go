package model

import (
	"fmt"
	"math"
	"time"
)

type BanType string

const (
	BanTypeAll       BanType = "all"
	BanTypeMic       BanType = "mic"  // not allowed to use mic; handle by banManager?
	BanTypeGame      BanType = "game" // not allow to use network services;not give session
	BanTypeAutomated BanType = "automated"
)

type BanCore struct {
	PlayerID string    `json:"player_id"`
	BanType  BanType   `json:"ban_type" gorm:"default:mic"`
	EndDate  time.Time `json:"end_date" gorm:"index" swaggertype:"string"`
}

type Ban struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	StartDate time.Time `json:"start_date" swaggertype:"string"`
	BanCore
}

//func (b *Ban) ToServiceBan() ServiceBan {
//	sb := ServiceBan{
//		//ID:             -1,
//		//Reason:        []byte{},
//		StartDatetime: b.StartDate,
//		ServiceBanCore: ServiceBanCore{
//			TargetType:  BanTargetPlatformId,
//			TargetId:    b.PlayerID,
//			BanType:     b.BanType,
//			EndDatetime: b.EndDate,
//		},
//	}
//
//	sb.FillReason(map[string]string{"from": "table bans"})
//	return sb
//}
//
//func (b *Ban) ToResponse() BanResponse {
//	return BanResponse{Ban: *b, StartDate: JsonTime{b.StartDate}, EndDate: JsonTime{b.EndDate}}
//}

type BanResponse struct {
	Ban

	//overwrite type
	StartDate JsonTime `json:"start_date" swaggertype:"string"`
	EndDate   JsonTime `json:"end_date" swaggertype:"string"`
}

type BanInfo struct {
	MicBanCount  int       `json:"mic" gorm:"column:mic"`
	GameBanCount int       `json:"game" gorm:"column:game"`
	Now          time.Time `json:"now"`
	EndDate      time.Time `json:"end_date" swaggertype:"string"`
}

//func IsPlayerGameBan(playerID string) (banned bool, err error) {
//	banned = true // default is banned
//	var count int64
//	result := myDB.DBConn.
//		Model(&Ban{}).
//		Where(&Ban{BanCore: BanCore{PlayerID: playerID, BanType: BanTypeGame}}).
//		Where("end_date > ? ", CurrentTimestamp()).
//		Count(&count)
//
//	if err = result.Error; err == nil {
//		banned = count != 0
//	}
//
//	return
//}
//
//func GetPlayerGameBan(playerID string) (ban Ban, err error) {
//	result := myDB.DBConn.
//		Where(&Ban{BanCore: BanCore{PlayerID: playerID, BanType: BanTypeGame}}).
//		Where("end_date > ? ", CurrentTimestamp()).
//		Last(&ban)
//
//	if err = result.Error; err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
//		err = nil
//	}
//
//	return
//}

//// GetPlayerUUIDsGameBan Retrieve a list of player_id that have been issued GameBan; NO MATTER the ban expired or not
//func GetPlayerUUIDsGameBan() (uuids []string, err error) {
//	SubQueryGetPlayerUUIDsGameBan := myDB.DBConn.
//		Model(&Ban{}).
//		Select("player_id").
//		Distinct("player_id").
//		Where(&Ban{BanCore: BanCore{BanType: BanTypeGame}})
//
//	result := myDB.DBConn.
//		Table("(?) as player_id", SubQueryGetPlayerUUIDsGameBan).
//		Find(&uuids)
//
//	if err = result.Error; err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
//		err = nil
//	}
//
//	return
//}

//func GetPlayerLastBan(playerID string) (ban Ban, err error) {
//	// $query = "SELECT * FROM bans WHERE player_id = :player_id AND end_date > CURRENT_TIMESTAMP() ORDER BY end_date DESC LIMIT 1";
//	result := myDB.DBConn.
//		Order("end_date DESC").
//		First(&ban, "player_id = ? AND end_date > ? ", playerID, CurrentTimestamp())
//
//	if err = result.Error; err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
//		err = nil
//	}
//
//	return
//}

//func GetPlayerBan(playerID string) (ban Ban, err error) {
//
//	//// $query = "SELECT id FROM bans WHERE player_id = :player_id AND CURRENT_TIMESTAMP() BETWEEN start_date AND end_date LIMIT 1";
//	result := myDB.DBConn.
//		Take(&ban, "player_id = ? AND ? BETWEEN start_date AND end_date", playerID, CurrentTimestamp())
//
//	if err = result.Error; err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
//		err = nil
//	}
//
//	return
//}

//func GetPlayerBans(playerID string, isTimed bool, interval int) (bans []Ban, err error) {
//	//$query = "SELECT id, ban_type, start_date, end_date
//	//	FROM bans
//	//	WHERE bans.player_id = :player_id";
//	var result *gorm.DB
//	if isTimed {
//		result = myDB.DBConn.Where("player_id = ? AND start_date >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? MINUTE)", playerID, interval).Find(&bans)
//	} else {
//		result = myDB.DBConn.Where(&Ban{BanCore: BanCore{PlayerID: playerID}}).Find(&bans)
//	}
//	err = result.Error
//
//	return
//}

// GetPlayerBanInfo get specific player's active ban count
//func GetPlayerBanInfo(playerID string) (binfo BanInfo, err error) {
//	//$query = "SELECT SUM(IF(ban_type = \"mic\", 1, 0)) as mic, SUM(IF(ban_type = \"game\", 1, 0)) as game , end_date
//	//FROM bans
//	//WHERE player_id = :player_id AND CURRENT_TIMESTAMP() BETWEEN start_date AND end_date";
//
//	//selectQuery := "SUM(IF(ban_type = 'mic', 1, 0)) as mic, SUM(IF(ban_type = 'game', 1, 0)) as game, end_date"
//	selectQuery := "SUM(IF(ban_type = 'mic', 1, 0)) as mic, SUM(IF(ban_type = 'game', 1, 0)) as game, UTC_TIMESTAMP() as now, end_date"
//	whereQuery := "player_id = ? AND ? BETWEEN start_date AND end_date"
//
//	//myDB.DBConn.Model(&Ban{}).Where(whereQuery, playerID, CurrentTimestamp()).Select(selectQuery).Group("end_date").Scan(&binfo)
//	result := myDB.DBConn.Model(&Ban{}).Where(whereQuery, playerID, CurrentTimestamp()).Select(selectQuery).Group("end_date").Scan(&binfo)
//
//	err = result.Error
//
//	return
//}

//func BanPlayerAdmin(playerId string, count int, banType BanType, isAutomated bool) (err error) {
//	//if ($is_automated) {
//	//	$query = "INSERT INTO `bans` (`player_id`, `ban_type`, `start_date`, `end_date`)
//	//	VALUES (:player_id, :ban_type, current_timestamp(), current_timestamp() + INTERVAL :duration MINUTE);";
//	//}
//	//else {
//	//	$query = "INSERT INTO `bans` (`player_id`, `ban_type`, `start_date`, `end_date`)
//	//	VALUES (:player_id, :ban_type, current_timestamp(), current_timestamp() + INTERVAL :duration DAY);";
//	//}
//
//	startTime := CurrentTimestamp()
//
//	var endTime time.Time
//	if isAutomated {
//		endTime = startTime.Add(time.Minute * time.Duration(count))
//	} else {
//		endTime = startTime.AddDate(0, 0, count) //now.Add(time.Hour * time.Duration(24*count))
//	}
//
//	ban := Ban{
//		StartDate: startTime,
//		BanCore:   BanCore{PlayerID: playerId, BanType: banType, EndDate: endTime},
//	}
//
//	result := myDB.DBConn.Create(&ban)
//	if result.Error != nil {
//		err = result.Error
//	}
//
//	return
//}

//func EndBanAdmin(ban_id uint64) (rowsAffected int64, err error) {
//	////$query = "UPDATE bans SET end_date = current_timestamp() WHERE id = :ban_id";
//
//	result := myDB.DBConn.Model(&Ban{}).Where("id = ?", ban_id)
//	if result.Update("end_date", CurrentTimestamp()); result.Error != nil {
//		err = result.Error
//	} else {
//		rowsAffected = result.RowsAffected
//	}
//	return
//}

func GetTimeDifferenceInHumanReadableFormat(startDate time.Time, endDate time.Time) (secondsUntilOver int, timeUntilOver string) {
	interval := endDate.Sub(startDate).Seconds()
	secondsUntilOver = int(interval)
	days := math.Floor(interval / 86400)
	interval = interval - (days * 86400)
	hours := math.Floor(interval / 3600)
	interval = interval - (hours * 3600)
	minutes := math.Floor(interval / 60)

	if days == 0 && hours == 0 && minutes == 0 {
		timeUntilOver = fmt.Sprintf("%d seconds", int(interval))
	} else if days == 0 && hours == 0 {
		timeUntilOver = fmt.Sprintf("%d minutes", int(minutes))
	} else if days == 0 {
		timeUntilOver = fmt.Sprintf("%d hours and %d minutes", int(hours), int(minutes))
	} else {
		timeUntilOver = fmt.Sprintf("%d days and %d hours and %d minutes", int(days), int(hours), int(minutes))
	}
	return
}

//// DuplicateGameBanInBanIntoServiceBan is used for migrate only GameBan from table `bans` to table `service_bans`
//func DuplicateGameBanInBanIntoServiceBan() {
//
//	//All GameBan weather expired or not; no mic ban
//	SubQueryGetPlayerUUIDsGameBan := myDB.DBConn.
//		Model(&Ban{}).
//		Where(&Ban{BanCore: BanCore{BanType: BanTypeGame}})
//
//	var gbs []Ban
//	result1 := myDB.DBConn.
//		Table("(?)as uids", SubQueryGetPlayerUUIDsGameBan).
//		Find(&gbs)
//	if result1.Error != nil || result1.RowsAffected <= 0 {
//		return
//	}
//
//	log.Printf("%v] %#v", len(gbs), gbs)
//
//	var sbs = make([]ServiceBan, len(gbs))
//	for idx, gb := range gbs {
//		sbs[idx] = gb.ToServiceBan()
//	}
//
//	log.Printf("%v] %#v", len(sbs), sbs)
//
//	myDB.DBConn.Create(&sbs)
//}

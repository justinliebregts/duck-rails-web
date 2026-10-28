package model

import (
	"encoding/json"
	"errors"
	myDB "gofiber/database/mysql"
	myERR "gofiber/errors"
	"math"
	"time"

	"gorm.io/gorm"
)

type BanTargetType string

const (
	BanTargetPlatformId BanTargetType = "platform"
	BanTargetDeviceUuid BanTargetType = "device"
)

var (
	emptyMap = map[string]string{}
)

type ServiceBanWithExpirationMark struct {
	ServiceBan
	Expired bool `json:"exp" gorm:"-"`
}

func NewServiceBanWithExpirationMarkSlice(sbs []ServiceBan) []ServiceBanWithExpirationMark {
	result := make([]ServiceBanWithExpirationMark, len(sbs))
	now := CurrentTimestamp()
	for i, sb := range sbs {
		result[i] = ServiceBanWithExpirationMark{
			ServiceBan: sb,
			Expired:    now.After(sb.EndDatetime),
		}
	}
	return result
}

type ServiceBan struct {
	ID            uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Reason        []byte    `json:"-" gorm:"default:{}"`
	ReasonStr     string    `json:"reason" gorm:"-"`
	StartDatetime time.Time `json:"start_datetime" swaggertype:"string"`
	ServiceBanCore
}

// ToBan let ServiceBan pretend to be Ban, should be gone in the future?
func (s *ServiceBan) ToBan() Ban {

	convertedID := uint(0)

	// Check for overflow
	if s.ID <= math.MaxUint {
		convertedID = uint(s.ID)
	}

	return Ban{ID: convertedID, StartDate: s.StartDatetime, BanCore: BanCore{PlayerID: s.TargetId, BanType: s.BanType, EndDate: s.EndDatetime}}
}

func (s *ServiceBan) ToResponse() BanResponse {
	return BanResponse{Ban: s.ToBan(), StartDate: JsonTime{s.StartDatetime}, EndDate: JsonTime{s.EndDatetime}}
}

func (s *ServiceBan) FillReason(reason interface{}) {
	if strReason, ok := reason.(string); ok {
		if len(strReason) > 2 {
			s.Reason = []byte(strReason)
			return
		}

		reason = emptyMap
	}

	// not in string, treat as json
	s.Reason, _ = json.Marshal(reason)
}

func (s *ServiceBan) FormReasonString() {
	s.ReasonStr = string(s.Reason)
}

type ServiceBanCore struct {
	TargetType  BanTargetType `json:"target_type" gorm:"type:enum('platform','device');default:platform"`
	TargetId    string        `json:"target_id" gorm:"default:platform_id_123"`
	BanType     BanType       `json:"ban_type" gorm:"default:mic"`
	EndDatetime time.Time     `json:"end_datetime" gorm:"index" swaggertype:"string"`
}

func GetTargetIDAnyGameBan(playerID string, deviceUuid string) (count int64, err error) {
	subQuery := myDB.DBConn.
		Model(&ServiceBan{}).
		Where(&ServiceBan{ServiceBanCore: ServiceBanCore{BanType: BanTypeGame}}).
		Where("end_datetime > ? ", CurrentTimestamp())

	result := myDB.DBConn.
		Table("(?) as u", subQuery).
		Where(&ServiceBan{ServiceBanCore: ServiceBanCore{TargetType: BanTargetPlatformId, TargetId: playerID}})

	if len(deviceUuid) > 0 && deviceUuid != "fake_device_uuid" {
		result.Or(&ServiceBan{ServiceBanCore: ServiceBanCore{TargetType: BanTargetDeviceUuid, TargetId: deviceUuid}})
	}
	result.Count(&count)

	err = result.Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}

	return
}

func GetSpecificServiceBanInfoBySession(base SessionBase, nonExpiredOnly bool) (sbs []ServiceBan, err error) {
	var pl Player
	if pl, err = TryGetPlayerFromDB(&base); err != nil {
		return GetSpecificServiceBanInfo(BanTypeAll, base.ComposedID(), "", nonExpiredOnly)
	}

	return GetSpecificServiceBanInfo(BanTypeAll, pl.ID, pl.DeviceUUID, nonExpiredOnly)
}

// GetSpecificActiveServiceBanInfo is used to fetch Ban for OTHER THAN certain condition
// it only retrieve active records for player ban determination
func GetSpecificActiveServiceBanInfo(banType BanType, playerUuid string, deviceUuid string) (sbs []ServiceBan, err error) {
	return GetSpecificServiceBanInfo(banType, playerUuid, deviceUuid, true)
}

// GetSpecificServiceBanInfo is used to fetch Ban for certain condition
// `target_type` = 'device' AND `target_id` = 'assigned_device_id' AND
// `target_type` = 'platform' AND “target_id` = 'assign_player_uuid'
func GetSpecificServiceBanInfo(banType BanType, playerUuid string, deviceUuid string, nonExpiredOnly bool) (sbs []ServiceBan, err error) {

	subQ := myDB.DBConn.
		Model(&ServiceBan{})

	if nonExpiredOnly {
		subQ = subQ.Where("end_datetime >= ?", CurrentTimestamp())
	}

	switch banType {
	case BanTypeGame:
		subQ.Where("ban_type = ?", string(BanTypeGame))
	case BanTypeMic:
		subQ.Where("ban_type = ?", string(BanTypeMic))
	default:
		// do nothing to match all BanType
	}

	// Construct the base query
	result := myDB.DBConn.Table("(?) as valid", subQ)

	// Check if searching for a platform ban
	searchForPlatformBan := false
	if len(playerUuid) > 0 {
		searchForPlatformBan = true
	}

	// Check if searching for a device ban
	searchForDeviceBan := false
	if len(deviceUuid) > 0 {
		searchForDeviceBan = true
	}

	if searchForPlatformBan && searchForDeviceBan {
		//result.Where("((target_type = ? AND target_id = ?) OR (target_type = ? AND target_id = ?))", BanTargetDeviceUuid, deviceUuid, BanTargetPlatformId, playerUuid)
		result.Where(&ServiceBan{ServiceBanCore: ServiceBanCore{TargetType: BanTargetDeviceUuid, TargetId: deviceUuid}}).
			Or(&ServiceBan{ServiceBanCore: ServiceBanCore{TargetType: BanTargetPlatformId, TargetId: playerUuid}})
	} else if !searchForPlatformBan && searchForDeviceBan {
		result.Where(&ServiceBan{ServiceBanCore: ServiceBanCore{TargetType: BanTargetDeviceUuid, TargetId: deviceUuid}})
	} else if searchForPlatformBan && !searchForDeviceBan {
		result.Where(&ServiceBan{ServiceBanCore: ServiceBanCore{TargetType: BanTargetPlatformId, TargetId: playerUuid}})
	}

	if !searchForDeviceBan && !searchForPlatformBan {
		return []ServiceBan{}, myERR.NoPlayerMatch
	}

	err = result.Scan(&sbs).Error
	if err != nil || sbs == nil {
		sbs = []ServiceBan{}
	}

	for _, sb := range sbs {
		sb.FormReasonString()
	}

	return
}

func GetPlayerServiceBanInfo(playerUUID string, deviceID string) (binfo BanInfo, err error) {

	selectQuery := `SUM(IF(ban_type = 'mic', 1, 0)) as mic,
SUM(IF(ban_type = 'game', 1, 0)) as game,
UTC_TIMESTAMP() as now, end_datetime as end_date`

	whereQuery := `target_id = ? AND ? BETWEEN start_datetime AND end_datetime`
	orQuery := `target_id = ? AND ? BETWEEN start_datetime AND end_datetime`

	result := myDB.DBConn.
		Model(&ServiceBan{}).
		Where(whereQuery, playerUUID, CurrentTimestamp())

	if len(deviceID) > 0 {
		result.Or(orQuery, deviceID, CurrentTimestamp())
	}

	result.Select(selectQuery).
		Group("end_datetime").
		Scan(&binfo)

	err = result.Error

	return
}

type ServiceBanAdmin struct {
	ServiceBan
	PlayerName       string `json:"player_name"`
	PlayerIP         string `json:"player_ip"`
	DeviceUUID       string `json:"device_uuid"`
	DurationSeconds  int64  `json:"duration_seconds"`
	SecondsRemaining int64  `json:"seconds_remaining"`
	TimeRemaining    string `json:"time_remaining" gorm:"-"`
}

var (
	allServiceBansAdminQry = `service_bans.id, target_type, target_id, players.name as player_name, players.ip as player_ip,
COALESCE(players.device_uuid, IF(service_bans.target_type='device', service_bans.target_id, NULL)) as device_uuid,
ban_type, start_datetime, end_datetime,
TIMESTAMPDIFF(SECOND, start_datetime, end_datetime) as duration_seconds,
TIMESTAMPDIFF(SECOND, ?, end_datetime) as seconds_remaining`
	allServiceBansAdminLjoin = "left join players on service_bans.target_type='platform' AND players.id = service_bans.target_id"

	searchCondNonID  = "service_bans.target_id LIKE ? OR  players.name LIKE ? OR players.ip LIKE ?"
	searchCondWithID = "service_bans.target_id LIKE ? OR  players.name LIKE ? OR players.ip LIKE ? OR service_bans.id = ?"
)

// func GetAllServiceBansAdmin(activeOnly bool, banType BanType) (sbAdmins []ServiceBanAdmin, err error) {
// 	now := CurrentTimestamp()

// 	result := myDB.DBConn.
// 		Model(&ServiceBan{}).
// 		Select(allServiceBansAdminQry, now).
// 		Joins(allServiceBansAdminLjoin).
// 		Order("service_bans.id DESC")

// 	if activeOnly {
// 		result = result.Where("end_datetime > ?", now)
// 	}
// 	if banType != "" && banType != BanTypeAll {
// 		result = result.Where("ban_type = ?", banType)
// 	}

// 	err = result.Scan(&sbAdmins).Error
// 	if result.RowsAffected == 0 {
// 		sbAdmins = []ServiceBanAdmin{}
// 	}

// 	return
// }

func CountServiceBansAdminWithQuery(ts time.Time, q *ListQuery, activeOnly bool, banType BanType) (total int64, err error) {

	result := myDB.DBConn.
		Model(&ServiceBan{}).
		//Select(allServiceBansAdminQry, ts).
		Joins(allServiceBansAdminLjoin)

	if activeOnly {
		result = result.Where("end_datetime > ?", ts)
	}
	if banType != "" && banType != BanTypeAll {
		result = result.Where("ban_type = ?", banType)
	}

	if q != nil {
		if ok, searchPhase := q.ShouldSearch(); ok {
			searchFuzzy := "%" + searchPhase + "%"

			if sid, lerr := q.SearchAsID(); lerr != nil {
				result = result.Where(searchCondNonID, searchFuzzy, searchFuzzy, searchFuzzy)
			} else {
				result = result.Where(searchCondWithID, searchFuzzy, searchFuzzy, searchFuzzy, sid)
			}
		}
	}

	err = result.Count(&total).Error

	return
}

func GetAllServiceBansAdminWithQuery(ts time.Time, q *ListQuery, activeOnly bool, banType BanType) (sbAdmins []ServiceBanAdmin, err error) {

	result := myDB.DBConn.
		Model(&ServiceBan{}).
		Select(allServiceBansAdminQry, ts).
		Joins(allServiceBansAdminLjoin)

	if activeOnly {
		result = result.Where("end_datetime > ?", ts)
	}
	if banType != "" && banType != BanTypeAll {
		result = result.Where("ban_type = ?", banType)
	}

	limit := 50
	offset := 0

	if q != nil {
		if q.PagePerCount > 0 && q.PagePerCount <= 100 {
			limit = q.PagePerCount
		}
		offset = q.Offset()

		if ok, searchPhase := q.ShouldSearch(); ok {
			searchFuzzy := "%" + searchPhase + "%"

			if sid, lerr := q.SearchAsID(); lerr != nil {
				result = result.Where(searchCondNonID, searchFuzzy, searchFuzzy, searchFuzzy)
			} else {
				result = result.Where(searchCondWithID, searchFuzzy, searchFuzzy, searchFuzzy, sid)
			}
		}
	}

	err = result.
		Order("service_bans.id DESC").
		Limit(limit).
		Offset(offset).
		Scan(&sbAdmins).Error
	if result.RowsAffected == 0 {
		sbAdmins = []ServiceBanAdmin{}
	}

	return
}

func BanServiceAdminWithDuration(targetType BanTargetType, targetID string, duration time.Duration, banType BanType, reason interface{}) (id uint64, err error) {

	startTime := CurrentTimestamp()
	endTime := startTime.Add(duration)

	if endTime.After(time.Date(2038, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return 0, errors.New("endTime exceeds 2038 limit for TIMESTAMP")
	}

	ban := ServiceBan{
		StartDatetime:  startTime,
		ServiceBanCore: ServiceBanCore{TargetType: targetType, TargetId: targetID, BanType: banType, EndDatetime: endTime},
	}

	ban.FillReason(reason)

	result := myDB.DBConn.
		Create(&ban)
	if err = result.Error; err == nil {
		id = ban.ID
	}

	return
}

func BanServiceAdmin(targetType BanTargetType, targetID string, count int, banType BanType, isAutomated bool, reason interface{}) (id uint64, err error) {

	unit := time.Hour * 24 // Default to days
	if isAutomated {
		unit = time.Minute // Change to minutes if automated
	}

	//calculate duration
	dur := unit * time.Duration(count)

	return BanServiceAdminWithDuration(
		targetType, targetID,
		dur, // calculated duration
		banType, reason)
}

func EndBanServiceAdmin(serviceBanId uint64) (rowsAffected int64, err error) {

	sb := ServiceBan{ID: serviceBanId, ServiceBanCore: ServiceBanCore{EndDatetime: CurrentTimestamp()}}

	result := myDB.DBConn.
		Model(&sb).
		Updates(&sb)

	if err = result.Error; err == nil {
		rowsAffected = result.RowsAffected
	}
	return
}

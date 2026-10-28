package model

import (
	myDB "gofiber/database/mysql"
	"time"
)

type PlatformSessionEntry struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `gorm:"not null;default:UTC_TIMESTAMP"`
	SessionBase
}

func (p *PlatformSessionEntry) TableName() string {
	return "platform_session_entries"
}

type PlatformCount struct {
	Platform string `json:"platform"`
	Count    int    `json:"count"`
}

func (p *PlatformCount) IsDevice() bool {
	switch p.Platform {
	case "ROOMBOT", "DEBUGGER":
		return false
	}

	return true
}

type PlatformRatio struct {
	PlatformCount
	Count float32 `json:"rate"`
}

func GetPlatformUsageStatsFromDateRange(start time.Time, end time.Time, distinctByPlayer bool) (pcs []PlatformCount, err error) {
	result := myDB.DBConn.Model(&PlatformSessionEntry{})
	if distinctByPlayer {
		result.Select("platform, count(distinct player_id) as count")
	} else {
		result.Select("platform, count(player_id) as count")
	}

	result.Group("platform").
		Where("created_at  BETWEEN ? AND ?", start, end).
		Scan(&pcs)
	err = result.Error
	if result.RowsAffected == 0 {
		pcs = []PlatformCount{}
	}

	return
}

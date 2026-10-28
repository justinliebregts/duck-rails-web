package model

import (
	"time"
)

type AccessRecord struct {
	//we wanted to store in a log the player id, their device uuid, ip, and the time,
	PlayerUUID  string    `json:"usr_id"`
	DeviceID    string    `json:"dev_id"`
	DisplayName string    `json:"dsp_nme"`
	IPAddress   []string  `json:"ip"`
	Datetime    time.Time `json:"timestamp"`
	AppSource   string    `json:"app_source"`
}

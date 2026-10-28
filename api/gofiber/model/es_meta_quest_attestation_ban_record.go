package model

type ActionType string

const (
	ActionBan   ActionType = "BAN"
	ActionUnban ActionType = "UNBAN"
)

type ESDeviceBanRecord struct {
	PlatformID       string     `json:"pfm_id"`
	Action           ActionType `json:"action"`
	UniqueId         string     `json:"unique_id"`
	BanID            string     `json:"ban_id"`
	RemainingBanTime int        `json:"remaining_time_in_minute" validate:"gte=0,lte=52560000"`
	Reason           string     `json:"ban_reason"`
	Timestamp        string     `json:"timestamp"`
}

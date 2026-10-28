package model

import (
	"database/sql"
	"encoding/json"
	mySTO "gofiber/cache"
	"net"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

const (
	playerAlivenessCacheIndex = 3
	keyPlayerAlivenessPrefix  = "online/"
	filterPlayerAliveness     = "online/*"
)

var (
	BanForHashVersion = time.Hour * 24 * 365 * 5

	HeartbeatFrequency = time.Second * 10
	HeartbeatLifeCycle = time.Second * 15
)

// RetrieveOnlinePlayerAsCacheKeys retrieve online/* from Cache
func RetrieveOnlinePlayerAsCacheKeys() ([]string, error) {
	return mySTO.ScanWithDBIndex(playerAlivenessCacheIndex, filterPlayerAliveness)
}

func RetrieveOnlinePlayersAsPlayerUUIDs() (playerUUIDs []string, err error) {
	var keys []string
	if keys, err = RetrieveOnlinePlayerAsCacheKeys(); err != nil {
		return
	}

	playerUUIDs = make([]string, len(keys))
	for idx, val := range keys {
		playerUUIDs[idx] = strings.TrimPrefix(val, keyPlayerAlivenessPrefix)
	}

	return
}

// CreateOnlinePlayerToCache is PUT online/gr_uuid with content
func CreateOnlinePlayerToCache(cacheKey string, content []byte) error {
	return mySTO.CreateWithDBIndex(playerAlivenessCacheIndex, cacheKey, content, HeartbeatLifeCycle)
}

func RemoveOnlinePlayerFromCache(cacheKey string) error {
	return mySTO.DeleteWithDBIndex(playerAlivenessCacheIndex, cacheKey)
}

func IsOnlinePlayerInCache(cacheKey string) (exist bool, err error) {
	return mySTO.ExistWithDBIndex(playerAlivenessCacheIndex, cacheKey)
}

// LoadOnlinePlayerContentFromCache is GET content for online/gr_uuid
func LoadOnlinePlayerContentFromCache(cacheKey string) ([]byte, error) {
	return mySTO.FetchWithDBIndex(playerAlivenessCacheIndex, cacheKey)
}

// Player is used in GORM/ DB
type Player struct {
	ID          string       `json:"id" gorm:"type:varchar(48) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;primaryKey"`
	Name        string       `json:"name" gorm:"type:varchar(128);"`
	Platform    string       `json:"platform" gorm:"type:enum('OCULUS','STEAM','IQIYI','EMAIL','PICO','APPLE','GOOGLE','ROOMBOT','DEBUGGER','YVR');"`
	IP          string       `json:"ip"`
	IPs         []string     `json:"-" gorm:"-"`
	Region      string       `json:"region"`
	Version     string       `json:"version" gorm:"type:varchar(16);"`
	Device      string       `json:"device" gorm:"type:varchar(128);"`
	DeviceUUID  string       `json:"device_uuid" gorm:"type:varchar(128);"`
	IsConverted bool         `json:"is_converted" gorm:"is_converted"`
	CreatedAt   JsonNullTime `json:"created" gorm:"column:created" swaggertype:"string"`
	UpdatedAt   JsonNullTime `json:"updated" gorm:"column:updated" swaggertype:"string"`
}

func (p *Player) Update(f *PlayerRegistrationForm) {
	//Updating "Volatile" data
	p.Name = f.Name
	p.Region = f.Region
	p.Version = f.Version
	p.IPs = f.IPs
	p.Device = f.DeviceName
	p.DeviceUUID = f.DeviceUuid
	p.UpdatedAt = JsonNullTime{
		NullTime: sql.NullTime{
			Time:  CurrentTimestamp(),
			Valid: true,
		},
	}
}

// FillIPFromIPs stores one canonical address, never a forwarded proxy chain.
// Forwarded addresses retain their existing client-first ordering.
func (p *Player) FillIPFromIPs() {
	for _, value := range p.IPs {
		for _, candidate := range strings.Split(value, ",") {
			if ip := net.ParseIP(strings.TrimSpace(candidate)); ip != nil {
				p.IP = ip.String()
				return
			}
		}
	}
}

func (p *Player) ToMap() (m fiber.Map) {

	bytes, _ := json.Marshal(p)
	json.Unmarshal(bytes, &m)

	//delete(m, "created")
	//delete(m, "updated")
	delete(m, "is_converted")

	return m
}

func NewPlayer(form *PlayerRegistrationForm) Player {
	return Player{
		ID:         form.ComposedID(), //model.CombinedKey(p.PlayerID, p.Platform),
		Platform:   form.Platform,
		Name:       form.Name,
		Region:     form.Region,
		Version:    form.Version,
		IPs:        form.IPs,
		Device:     form.DeviceName,
		DeviceUUID: form.DeviceUuid,
		//IsConverted bool         `json:"is_converted"`
		CreatedAt: JsonNullTime{sql.NullTime{Time: CurrentTimestamp(), Valid: true}},
		// UpdatedAt: model.JsonNullTime `json:"updated"`
	}
}

type HeartbeatForm struct {
	AppVersion              string `form:"ver" json:"ver"`
	Region                  string `form:"reg" json:"reg" validate:"required"`
	Lobby                   string `form:"lob" json:"lob" validate:"required"`
	Room                    string `form:"rom" json:"rom" validate:"required"`
	Mode                    string `form:"mod" json:"mod" validate:"required"`
	Difficulty              string `form:"dif" json:"dif" validate:"required"`
	Map                     string `form:"map" json:"map" validate:"required"`
	PlayerCount             string `form:"plc" json:"plc" validate:"required"`
	PrivateRoom             string `form:"pri" json:"pri" validate:"required"`
	AllowFriendsWithoutCode string `form:"awc" json:"awc" validate:"required"`
	Visibility              string `form:"vis" json:"vis"`
}

func (p *HeartbeatForm) Validate() (errMsgs []string) {
	if err := validate.Struct(p); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

func (p *HeartbeatForm) ConsideredAsOffline() bool {
	return strings.EqualFold(p.Visibility, "offline")
}

func (p *HeartbeatForm) ConsideredAsOnline() bool {
	return !p.ConsideredAsOffline()
}

func (p *HeartbeatForm) All() []string {
	playerUUIDs, _ := RetrieveOnlinePlayersAsPlayerUUIDs()
	return playerUUIDs
}

func CacheKeyForAliveness(playerUUID string) string {
	return keyPlayerAlivenessPrefix + playerUUID
}
func (p *HeartbeatForm) buildCacheKey(playerUUID string) string {
	return CacheKeyForAliveness(playerUUID)
}

func (p *HeartbeatForm) ExistInCache(playerUUID string) (exist bool, err error) {
	return IsOnlinePlayerInCache(p.buildCacheKey(playerUUID))
}

// SaveToCache temp cache as player online
func (p *HeartbeatForm) SaveToCache(playerUUID string) (err error) {

	var by []byte
	if by, err = json.Marshal(p); err != nil {
		return
	}

	return CreateOnlinePlayerToCache(p.buildCacheKey(playerUUID), by)
}

func (p *HeartbeatForm) LoadFromCache(playerUUID string) (err error) {
	var byts []byte
	if byts, err = LoadOnlinePlayerContentFromCache(p.buildCacheKey(playerUUID)); err == nil {
		//log.Println(err.Error())
		err = json.Unmarshal(byts, p)
	}

	return
}

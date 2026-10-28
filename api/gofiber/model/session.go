package model

import (
	"encoding/json"
	"errors"
	mySTO "gofiber/cache"
	myDB "gofiber/database/mysql"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2/utils"
)

const (
	KeyPrefix                   = "session/"
	KeyPlayerPrefix             = "session/player/"
	KeyRefreshPrefix            = "session/refresh/"
	AccessTokenExpiry           = 20 * time.Minute // 1200 * 1000 = 20min
	AccessTokenExpiryEnlongated = 30 * time.Minute
	RefreshTokenExpiry          = 14 * 24 * time.Hour // 1209600 = 14 days

	recordSessionCreation = true
)

func buildCacheKey(aTkn string) string {
	return KeyPrefix + aTkn
}

func buildCacheKeyUnderUUID(uuid string) string {
	return KeyPlayerPrefix + uuid
}

func buildKeyRefresh(rTkn string) string {
	return KeyRefreshPrefix + rTkn
}

func CachePlayerUUIDKey(sb *SessionBase) string {
	if sb == nil {
		return ""
	}
	return buildCacheKeyUnderUUID(sb.ComposedID()) //KeyPlayerPrefix + ses.ComposedID()
}

func CacheAccTknKey(accTkn string) string {
	return buildCacheKey(accTkn)
}

func CacheRefTknKey(refTkn string) string {
	return buildKeyRefresh(refTkn)
}

type SessionRefresh struct {
	RefreshToken string `form:"refresh_token" json:"refresh_token" validate:"required,min=8"`
}

func (p *SessionRefresh) Validate() (errMsgs []string) {
	if err := validate.Struct(p); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

type Session struct {
	SessionBase
	AccessToken string `json:"access_token,omitempty" validate:"required,min=8"`
	SessionRefresh
}

func (ses *Session) Identity() SessionBase {
	return SessionBase{PlayerID: ses.PlayerID, Platform: ses.Platform}
}

func (ses *Session) IdentityToBytes() ([]byte, error) {

	return json.Marshal(ses.Identity())
}

func (ses *Session) Copy(emptyAccessToken bool) (sess Session) {
	if !emptyAccessToken {
		sess.AccessToken = ses.AccessToken
	}

	sess.Platform = ses.Platform
	sess.PlayerID = ses.PlayerID
	sess.RefreshToken = ses.RefreshToken
	return
}

func (ses *Session) Validate() (errMsgs []string) {
	if err := validate.Struct(ses); err != nil {
		for _, er := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, er.Error())
		}
	}
	return
}

func (ses *Session) ToBytes(emptyAccessToken bool) ([]byte, error) {
	sess := ses.Copy(emptyAccessToken)
	return json.Marshal(sess)
}

func (ses *Session) AutoFill() {

	if len(ses.AccessToken) <= 0 {
		ses.AccessToken = utils.UUIDv4()
	}

	if len(ses.RefreshToken) <= 0 {
		ses.RefreshToken = utils.UUIDv4()
	}

	//if config.UseB64 {
	//ses.AccessToken =
	//	b64.StdEncoding.EncodeToString([]byte(ses.AccessToken))
	//ses.RefreshToken =
	//	b64.StdEncoding.EncodeToString([]byte(ses.RefreshToken))
	//}
}

func (ses *Session) RetrievePlayerFromDB() (Player, error) {
	return RetrieveFromDB(ses.ComposedID())
}

func (ses *Session) Exist() (exist bool, err error) {

	var sesRetrieved Session
	sesRetrieved, err = RetrievePlayerCache(ses.AccessToken)
	if err != nil {
		// do nothing: err assigned
	} else if ses.PlayerID != sesRetrieved.PlayerID {
		err = errors.New("PlayerIDNotMatch")
		//} else if ses.Platform != sesRetrieved.Platform {
		//	err = errors.New("PlatformNotMatch")
	} else {
		exist = true
	}

	return
}

func (ses *Session) SaveToCache() (err error) {
	err = ses.SaveToCacheForDuration(AccessTokenExpiryEnlongated)
	return
}

func (ses *Session) SaveToCacheForDuration(dur time.Duration) (err error) {

	var bytes []byte
	if bytes, err = ses.ToBytes(true); err != nil {
		return
	}
	err = mySTO.Create(buildCacheKey(ses.AccessToken), bytes, dur)

	if recordSessionCreation {
		_, _ = CreatePlatformSessionEntry(&ses.SessionBase)
	}

	err = ses.SaveUnderPlayerID()
	err = ses.SaveRefreshToCache()
	return
}

func (ses *Session) SaveUnderPlayerIDPersistent() (err error) {
	return mySTO.Create(CachePlayerUUIDKey(&ses.SessionBase), []byte(ses.AccessToken), 0)
}

func (ses *Session) SaveUnderPlayerID() (err error) {

	err = mySTO.Create(CachePlayerUUIDKey(&ses.SessionBase), []byte(ses.AccessToken), AccessTokenExpiryEnlongated)
	return
}

func (ses *Session) SaveRefreshToCache() (err error) {
	var res []byte
	if res, err = json.Marshal(ses.Identity()); err != nil {
		return
	}

	return mySTO.Create(buildKeyRefresh(ses.RefreshToken), res, RefreshTokenExpiry)
}

func BuildSessionFromSessionBase(sb SessionBase) (ses Session) {
	ses.SessionBase = sb

	ses.AutoFill()
	return
}

func CreatePlayerCache(playerID string, platform string) (ses Session, err error) {

	sb := NewSessionBase(platform, playerID)
	ses = BuildSessionFromSessionBase(sb)
	err = ses.SaveToCache()

	return
}

func RetrievePlayerCache(accTkn string) (ses Session, err error) {
	var byts []byte

	ck := buildCacheKey(accTkn)
	//	log.Printf("RetrievePlayerCache on %#v", ck)
	if byts, err = mySTO.Fetch(ck); err != nil {
		return
	}

	err = json.Unmarshal(byts, &ses)
	ses.AccessToken = accTkn
	return
}

type SessionResponse struct {
	SessionId             int     `json:"session_id"`
	GrId                  string  `json:"gr_id"`
	PlayerIdSeparated     string  `json:"player_id_separated,omitempty"`
	DisplayName           string  `json:"display_name,omitempty"`
	Platform              string  `json:"platform,omitempty"`
	AccessTokenExpiresIn  float64 `json:"access_token_expires_in"`
	RefreshTokenExpiresIn float64 `json:"refresh_token_expires_in"`
	Session
}

func (ses *Session) FormResponse() SessionResponse {
	res := SessionResponse{
		Session:               *ses,
		SessionId:             -1,
		GrId:                  ses.ComposedID(),
		AccessTokenExpiresIn:  AccessTokenExpiry.Seconds(),
		RefreshTokenExpiresIn: RefreshTokenExpiry.Seconds(),
	}

	res.Session.PlayerID = ""
	res.Session.Platform = ""
	//log.Printf("%#v", ses)
	return res
}

func (ses *Session) FormResponseForRegister(dname string) SessionResponse {
	res := ses.FormResponse()
	res.DisplayName = dname
	res.PlayerIdSeparated = ses.PlayerID
	res.Platform = ses.Platform
	return res
}

// CreatePlatformSessionEntry deprecated, should use (s *playerServiceImpl) CreatePlatformSessionEntry
func CreatePlatformSessionEntry(sb *SessionBase) (rowAffected int64, err error) {

	entry := PlatformSessionEntry{
		SessionBase: *sb,
	}

	result := myDB.DBConn.Create(&entry)

	return result.RowsAffected, result.Error
}

// IsGameBanned deprecated, should use playerServiceImpl
func IsGameBanned(sb *SessionBase) (gamBanned bool, err error) {
	//return IsPlayerGameBan(sb.ComposedID())

	gamBanned = true
	var cnt int64
	if cnt, err = GetTargetIDAnyGameBan(sb.ComposedID(), ""); err == nil {
		if cnt <= 0 {
			gamBanned = false
			return
		}
	}

	return
}

//func (sb *SessionBase) GetLastGameBan() (Ban, error) {
//	return GetPlayerGameBan(sb.ComposedID())
//}

// AllSessions deprecated, should use playerServiceImpl
func AllSessions() []string {
	keys, _ := mySTO.Scan(KeyPlayerPrefix + "*")
	res := make([]string, len(keys))
	for idx, val := range keys {
		res[idx] = strings.TrimPrefix(val, KeyPlayerPrefix)
	}

	return res
}

// TryGetPlayerFromDB deprecated, should use playerServiceImpl
func TryGetPlayerFromDB(sb *SessionBase) (Player, error) {
	return SearchBySessionBase(*sb)
}

// TrySaveDisplayNameToCache deprecated, should use playerServiceImpl
func TrySaveDisplayNameToCache(sb *SessionBase, displayName string) error {
	return CreateDisplayNameToCache(sb.ComposedID(), displayName)
}

// CacheSavePlayerState deprecated, should use playerServiceImpl
func CacheSavePlayerState(sb *SessionBase, form HeartbeatForm) (recordInCache bool, err error) {
	recordInCache, _ = form.ExistInCache(sb.ComposedID())

	return recordInCache, form.SaveToCache(sb.ComposedID())
}

// CacheRemovePlayerState deprecated, should use playerServiceImpl
func CacheRemovePlayerState(sb *SessionBase) (err error) {
	return RemoveOnlinePlayerFromCache(sb.ComposedID())
}

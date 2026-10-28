package repository

import (
	"encoding/json"
	mySTO "gofiber/cache"
	myERR "gofiber/errors"
	myMDL "gofiber/model"
	"strings"
)

type Session interface {
	//CreateSessionTokens(sb *myMDL.SessionBase, recordSessionCreation bool, accTkn string, accTknContent SessionCache, refreshTknContent SessionCache) error
	//ExistAccTknBySessionCache(parsedSessionCache SessionCache) (bool, error)
	//ExistRefreshTknBySessionCache(parsedSessionCache SessionCache) (bool, error)
	//
	//DeleteRefreshToken(refreshTknContent SessionCache) error

	GetAccessTokenByUUID(myMDL.SessionBase) (accTkn string, err error)
	GetSession(accTkn string) (myMDL.Session, error)
	GetRefreshData(refTkn string) (myMDL.SessionBase, error)

	DeletePlayerMapping(myMDL.SessionBase) error
	DeleteAccessToken(accTkn string) error
	DeleteRefreshToken(refTkn string) error
}

type sessionRepositoryImpl struct {
}

func NewSession() Session {
	return &sessionRepositoryImpl{}
}

func (s *sessionRepositoryImpl) AllActivePlayerUUIDs() []string {
	// retrieve all data, starting from "session/player/" such as session/player/z|tester123
	keys, _ := mySTO.Scan(myMDL.KeyPlayerPrefix + "*")
	res := make([]string, len(keys))
	for idx, val := range keys {
		res[idx] = strings.TrimPrefix(val, myMDL.KeyPlayerPrefix)
	}

	return res
}

func (s *sessionRepositoryImpl) AllActivePlayerSessionBases() []*myMDL.SessionBase {
	// retrieve all data, starting from "session/player/" such as session/player/z|tester123
	keys, _ := mySTO.Scan(myMDL.KeyPlayerPrefix + "*")
	res := make([]*myMDL.SessionBase, len(keys))
	for idx, val := range keys {
		sb := myMDL.FormSessionBase(strings.TrimPrefix(val, myMDL.KeyPlayerPrefix))
		res[idx] = &sb
	}

	return res
}

func (s *sessionRepositoryImpl) GetAccessTokenByUUID(sb myMDL.SessionBase) (at string, err error) {
	var byts []byte

	ck := myMDL.CachePlayerUUIDKey(&sb)
	if byts, err = mySTO.Fetch(ck); err != nil {
		return
	}

	at = string(byts)
	//
	return
}

func (s *sessionRepositoryImpl) GetRefreshData(refTkn string) (sb myMDL.SessionBase, err error) {
	if len(refTkn) == 0 {
		err = myERR.NoFound
		return
	}

	var byts []byte

	ck := myMDL.CacheRefTknKey(refTkn)
	if byts, err = mySTO.Fetch(ck); err != nil {
		return
	}

	err = json.Unmarshal(byts, &sb)
	return
}

func (s *sessionRepositoryImpl) GetSession(accTkn string) (ses myMDL.Session, err error) {
	if len(accTkn) == 0 {
		err = myERR.NoFound
		return
	}

	var byts []byte

	ck := myMDL.CacheAccTknKey(accTkn)
	if byts, err = mySTO.Fetch(ck); err != nil {
		return
	}

	err = json.Unmarshal(byts, &ses)
	if err == nil {
		ses.AccessToken = accTkn
	}
	return
}

func (s *sessionRepositoryImpl) DeletePlayerMapping(sb myMDL.SessionBase) error {
	ck := myMDL.CachePlayerUUIDKey(&sb)
	return mySTO.Delete(ck)
}
func (s *sessionRepositoryImpl) DeleteAccessToken(accTkn string) error {
	ck := myMDL.CacheAccTknKey(accTkn)
	return mySTO.Delete(ck)
}
func (s *sessionRepositoryImpl) DeleteRefreshToken(refTkn string) error {
	ck := myMDL.CacheRefTknKey(refTkn)
	return mySTO.Delete(ck)
}

//func (s *sessionRepositoryImpl) ExistAccTknBySessionCache(parsedSessionCache SessionCache) (exist bool, err error) {
//	// use only SessionCache.JTI
//	return mySTO.Exist(parsedSessionCache.CacheAccTknKey())
//}
//
//func (s *sessionRepositoryImpl) ExistRefreshTknBySessionCache(parsedSessionCache SessionCache) (exist bool, err error) {
//	// use only SessionCache.JTI
//	return mySTO.Exist(parsedSessionCache.CacheRefreshKey())
//}
//
//type SessionCache struct {
//	UserID          string    `json:"user_id" example:"a|T:123456"`
//	JTI             string    `json:"jti" example:"1234-5678-1234567` // Unique JWT ID (for blacklisting)
//	IssuedAt        time.Time `json:"iat"`
//	Platform        string    `json:"pla" example:"apple"`        // Metadata not in the JWT
//	PlatformAccount string    `json:"pla_acc" example:"T:123456"` // Metadata not in the JWT
//}
//
//func (sc *SessionCache) FormSessionBase() myMDL.SessionBase {
//	return myMDL.FormSessionBase(sc.UserID)
//}
//
//func (sc *SessionCache) CacheAccTknKey() string {
//	return myMDL.KeyPrefix + sc.JTI
//}
//
//func (sc *SessionCache) CacheRefreshKey() string {
//	return myMDL.KeyRefreshPrefix + sc.JTI
//}
//
//func (sc *SessionCache) ToBytes() []byte {
//	b, _ := json.Marshal(sc)
//	return b
//}

//func (s *sessionRepositoryImpl) CreateSessionTokens(sb *myMDL.SessionBase, recordSessionCreation bool, accTkn string, accTknContent SessionCache, refreshTknContent SessionCache) (err error) {
//	//Access token; for token EXIST lookup
//	err = mySTO.Create(
//		accTknContent.CacheAccTknKey(), // session/{access_token.JTI}
//		accTknContent.ToBytes(),        // store accTknContent SessionCache
//		myMDL.AccessTokenExpiryEnlongated)
//
//	if recordSessionCreation {
//		_, _ = myMDL.CreatePlatformSessionEntry(sb)
//	}
//
//	// Secondary key, For OnlineCheck, paired with previous AT; content just access token
//	err = mySTO.Create(
//		myMDL.CachePlayerUUIDKey(sb), // session/player/{player_uuid}
//		[]byte(accTkn),               //  store accessToken
//		myMDL.AccessTokenExpiryEnlongated)
//
//	//Refresh token, EXIST lookup
//	err = mySTO.Create(
//		refreshTknContent.CacheRefreshKey(), // session/refresh/{refresh_token.JTI}
//		refreshTknContent.ToBytes(),         //store refreshTknContent SessionCache
//		myMDL.RefreshTokenExpiry)
//	return
//}

//func (s *sessionRepositoryImpl) DeleteRefreshToken(refreshTknContent SessionCache) error {
//	return mySTO.Delete(refreshTknContent.CacheRefreshKey()) // session/refresh/{refresh_token.JTI}
//}

//func (s *sessionRepositoryImpl) RetrieveRefresh(sr *myMDL.SessionRefresh) (ses myDOM.SessionBase, err error) {
//	var byts []byte
//
//	if byts, err = mySTO.Fetch(sr.CacheRefreshKey()); err != nil {
//		return
//	}
//
//	err = json.Unmarshal(byts, &ses)
//	return
//}
//
//func (s *sessionRepositoryImpl) Delete(sr *myMDL.SessionRefresh) error {
//	return mySTO.Delete(sr.CacheRefreshKey())
//}

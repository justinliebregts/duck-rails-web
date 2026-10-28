package service

import (
	myMDL "gofiber/model"
	myREPO "gofiber/repository"
)

type Session interface {
	RevokeFullSession(sb myMDL.SessionBase) error
}

type sessionServiceImpl struct {
	sesRepo myREPO.Session
}

func NewSessionService(sesRepo myREPO.Session) Session {
	return &sessionServiceImpl{sesRepo: sesRepo}
}

func (s *sessionServiceImpl) RevokeFullSession(sb myMDL.SessionBase) (err error) {
	var at string
	if at, err = s.sesRepo.GetAccessTokenByUUID(sb); err != nil {
		return err
	}

	var ses myMDL.Session
	if ses, err = s.sesRepo.GetSession(at); err != nil {
		return err
	}

	_ = s.sesRepo.DeletePlayerMapping(sb)
	_ = s.sesRepo.DeleteAccessToken(at)
	_ = s.sesRepo.DeleteRefreshToken(ses.AccessToken)

	return nil
}

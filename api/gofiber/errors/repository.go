package errors

import "errors"

var (
	Banned                  = errors.New("player banned")
	DeviceBanned            = errors.New("device banned")
	NoRowAffected           = errors.New("no row affected")
	NoRowCreated            = errors.New("no row created")
	NoRowFound              = errors.New("no row found")
	NoRowInput              = errors.New("no row input")
	NoTimeInput             = errors.New("no time input")
	NoRegisteredPlayerFound = errors.New("no registered player found")
	NoTeamFound             = errors.New("no team found")
	NoItemFound             = errors.New("no item found")
	NoFound                 = errors.New("no found")
	NoMatch                 = errors.New("no match")
	NoActionMatch           = errors.New("no action match")
	PlayerRegistered        = errors.New("player registered")
	NoPlayerMatch           = errors.New("no player match")
	ErrCannotOwnMultiple    = errors.New("cannot own multiple")
	ErrItemTypeNotAllowed   = errors.New("item type is not an avatar or weapon skin")
)

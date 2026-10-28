package errors

import "errors"

var (
	ErrSTOConnectionNotExist = errors.New("STO Connection NotExist")
	ErrNotInCache            = errors.New("NotInCache")
)

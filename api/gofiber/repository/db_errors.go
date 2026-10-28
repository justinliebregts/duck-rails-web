package repository

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	// MySQL 1062 / GORM duplicated key
	msg := err.Error()
	return errors.Is(err, gorm.ErrDuplicatedKey) ||
		strings.Contains(msg, "Duplicate entry") ||
		strings.Contains(msg, "1062")
}

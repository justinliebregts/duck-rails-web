package model

import (
	"gorm.io/gorm"
)

type IPBlacklist struct {
	gorm.Model
	IP     string `gorm:"type:varchar(45);uniqueIndex:idx_ip;not null"` // Supports IPv4 & IPv6
	Reason string `gorm:"type:varchar(255)"`
	//ExpiresAt *time.Time `gorm:"index"` // NULL means permanent ban
}

// func (b *IPBlacklist) IsExpired() bool {
// 	if b.ExpiresAt == nil {
// 		return false // Permanent
// 	}
// 	return time.Now().After(*b.ExpiresAt)
// }

package repository

import (
	"context"
	"gofiber/model"
	myMDL "gofiber/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IPCheckRepository interface {

	//BanIP(ctx context.Context, ip string, reason string, duration time.Duration) error
	BanIP(ctx context.Context, ip string, reason string) error

	List(ctx context.Context) ([]myMDL.IPBlacklist, error)
	IsBanned(ctx context.Context, ip string) (bool, error)

	UnbanIP(ctx context.Context, ip string) error
	RecoverIPBan(ctx context.Context, ip string) error
}

type ipCheckRepoImpl struct {
	dbReadWrite *gorm.DB
}

func NewIPCheckRepository(db *gorm.DB) IPCheckRepository {
	return &ipCheckRepoImpl{dbReadWrite: db}
}

func (r *ipCheckRepoImpl) List(ctx context.Context) ([]myMDL.IPBlacklist, error) {
	var bannedList []myMDL.IPBlacklist
	err := r.dbReadWrite.WithContext(ctx).
		Unscoped().       // list soft-deleted also
		Order("id DESC"). //
		Find(&bannedList).Error
	if err != nil {
		return nil, err
	}
	return bannedList, nil
}

func (r *ipCheckRepoImpl) IsBanned(ctx context.Context, ip string) (bool, error) {
	var count int64
	//now := time.Now()

	err := r.dbReadWrite.WithContext(ctx).
		Model(&myMDL.IPBlacklist{}).
		//Where("ip = ? AND (expires_at IS NULL OR expires_at > ?)", ip, now).
		Where("ip = ?", ip).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *ipCheckRepoImpl) BanIP(ctx context.Context, ip string, reason string) error {
	//var expiresAt *time.Time
	// if duration > 0 {
	// 	t := time.Now().Add(duration)
	// 	expiresAt = &t
	// }

	nw := time.Now()
	ban := myMDL.IPBlacklist{
		IP:     ip,
		Reason: reason,
		Model: gorm.Model{
			CreatedAt: nw,
			UpdatedAt: nw,
		},
		//ExpiresAt: expiresAt,
	}
	return r.dbReadWrite.
		WithContext(ctx).
		Unscoped().
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "ip"}}, // Targets unique constraint on the IP column
			DoUpdates: clause.Assignments(map[string]interface{}{
				"reason":     reason,
				"deleted_at": nil, // Clears soft-delete flag reliably in SQL
				"updated_at": nw,
			}),
		}).
		Create(&ban).Error
}

func (r *ipCheckRepoImpl) UnbanIP(ctx context.Context, ip string) error {
	return r.dbReadWrite.WithContext(ctx).
		Where("ip = ?", ip).
		Delete(&myMDL.IPBlacklist{}).Error
}

func (r *ipCheckRepoImpl) RecoverIPBan(ctx context.Context, ip string) error {
	return r.dbReadWrite.WithContext(ctx).
		Unscoped().
		Model(&model.IPBlacklist{}).
		Where("ip = ? AND deleted_at IS NOT NULL", ip).
		Update("deleted_at", nil).Error
}

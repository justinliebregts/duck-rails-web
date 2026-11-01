package duckz

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	myERR "gofiber/errors"
	myMDL "gofiber/model"
	dzMDL "gofiber/model/duckz"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EconomyRepository reads Duckz economy state. Writes that touch the coin balance go through WithTx.
type EconomyRepository interface {
	Balance(playerID string) (int, error)
	GetDepositByRequestID(requestID string) (*dzMDL.CoinDeposit, error)
	GetPurchaseByRequestID(requestID string) (*dzMDL.StorePurchase, error)
	ListPurchasedKeys(playerID string) ([]string, error)
	ListEntitlementSkus(playerID string) ([]string, error)
	CreateEntitlement(row *dzMDL.Entitlement) error
	ListAdjustments(playerID string, limit int) ([]dzMDL.CoinAdjustment, error)
	ListTransactions(playerID string, limit int) ([]dzMDL.AdminTransaction, error)
	// GetLoadoutSkin returns the equipped skin key; found is false if the player never saved one.
	GetLoadoutSkin(playerID string) (skinKey string, found bool, err error)
	SaveLoadoutSkin(playerID, skinKey string, now time.Time) error

	// WithTx runs fn in one DB transaction holding a row lock on the player.
	WithTx(playerID string, fn func(tx EconomyTx) error) error
}

// EconomyTx is the set of writes allowed while the player row is locked.
type EconomyTx interface {
	Balance(playerID string) (int, error)
	SumDepositCoinsSince(playerID string, since time.Time) (int, error)
	CountPurchases(playerID, itemKey string) (int64, error)
	CreateTransaction(tr *myMDL.Transaction) error
	CreateDeposit(row *dzMDL.CoinDeposit) error
	CreatePurchase(row *dzMDL.StorePurchase) error
}

type economyRepositoryImpl struct {
	dbc *gorm.DB
}

func NewEconomyRepository(dbc *gorm.DB) EconomyRepository {
	return &economyRepositoryImpl{dbc: dbc}
}

func balance(db *gorm.DB, playerID string) (bal int, err error) {
	err = db.Model(&myMDL.Transaction{}).
		Where("player_id = ? AND is_deleted = false", playerID).
		Select("COALESCE(SUM(credit) - SUM(debit), 0)").
		Scan(&bal).Error
	return
}

func (r *economyRepositoryImpl) Balance(playerID string) (int, error) {
	var coins int
	err := r.WithTx(playerID, func(tx EconomyTx) error { var err error; coins, err = tx.Balance(playerID); return err })
	return coins, err
}

func (r *economyRepositoryImpl) GetDepositByRequestID(requestID string) (*dzMDL.CoinDeposit, error) {
	var row dzMDL.CoinDeposit
	err := r.dbc.Where("request_id = ?", requestID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, myERR.NoMatch
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *economyRepositoryImpl) GetPurchaseByRequestID(requestID string) (*dzMDL.StorePurchase, error) {
	var row dzMDL.StorePurchase
	err := r.dbc.Where("request_id = ?", requestID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, myERR.NoMatch
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *economyRepositoryImpl) ListPurchasedKeys(playerID string) (keys []string, err error) {
	err = r.dbc.Model(&dzMDL.StorePurchase{}).
		Where("player_id = ?", playerID).
		Distinct().
		Pluck("item_key", &keys).Error
	if keys == nil {
		keys = []string{}
	}
	return
}

func (r *economyRepositoryImpl) ListTransactions(playerID string, limit int) (rows []dzMDL.AdminTransaction, err error) {
	err = r.dbc.Model(&myMDL.Transaction{}).
		Select("id, COALESCE(platform, '') AS platform, COALESCE(platform_account, '') AS platform_account, "+
			"COALESCE(platform_txn_id, '') AS platform_txn_id, debit, credit, is_deleted, date").
		Where("player_id = ?", playerID).
		Order("id DESC").
		Limit(limit).
		Scan(&rows).Error
	if rows == nil {
		rows = []dzMDL.AdminTransaction{}
	}
	return
}

func (r *economyRepositoryImpl) GetLoadoutSkin(playerID string) (string, bool, error) {
	var row dzMDL.Loadout
	res := r.dbc.Where("player_id = ?", playerID).Limit(1).Find(&row)
	if res.Error != nil {
		return "", false, res.Error
	}
	return row.SkinKey, res.RowsAffected > 0, nil
}

func (r *economyRepositoryImpl) SaveLoadoutSkin(playerID, skinKey string, now time.Time) error {
	return r.dbc.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "player_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"skin_key", "updated_at"}),
	}).Create(&dzMDL.Loadout{PlayerID: playerID, SkinKey: skinKey, UpdatedAt: now}).Error
}

func (r *economyRepositoryImpl) ListEntitlementSkus(playerID string) (skus []string, err error) {
	err = r.dbc.Model(&dzMDL.Entitlement{}).
		Where("player_id = ?", playerID).
		Pluck("sku", &skus).Error
	if skus == nil {
		skus = []string{}
	}
	return
}

func (r *economyRepositoryImpl) CreateEntitlement(row *dzMDL.Entitlement) error {
	return mapDuplicate(r.dbc.Create(row).Error)
}

func (r *economyRepositoryImpl) ListAdjustments(playerID string, limit int) (rows []dzMDL.CoinAdjustment, err error) {
	err = r.dbc.Model(&myMDL.Transaction{}).
		Select("id, COALESCE(platform_account, '') AS admin_id, credit - debit AS amount, date AS created_at").
		Where("player_id = ? AND platform = ? AND is_deleted = false", playerID, dzMDL.TransactionPlatformAdminAdjust).
		Order("id DESC").
		Limit(limit).
		Scan(&rows).Error
	if rows == nil {
		rows = []dzMDL.CoinAdjustment{}
	}
	return
}

func (r *economyRepositoryImpl) WithTx(playerID string, fn func(tx EconomyTx) error) error {
	return r.dbc.Transaction(func(tx *gorm.DB) error {
		var player myMDL.Player
		res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id").
			Where("id = ?", playerID).
			Limit(1).
			Find(&player)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return myERR.ErrDuckzPlayerNotFound
		}
		// Serialize the welcome grant with purchases and admin adjustments.
		// The receipt survives transaction deletion and a zero balance.
		var count int64
		if err := tx.Table("rails_wallet_grants").Where("player_id = ?", playerID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := tx.Exec("INSERT INTO rails_wallet_grants (player_id, created_at) VALUES (?, ?)", playerID, time.Now().UTC()).Error; err != nil {
				return err
			}
			grant := myMDL.Transaction{TransactionCreation: myMDL.TransactionCreation{PlayerID: playerID, Credit: 100,
				Platform: sql.NullString{String: "rails_welcome", Valid: true}}}
			if err := tx.Create(&grant).Error; err != nil {
				return err
			}
		}
		return fn(&economyTxImpl{tx: tx})
	})
}

type economyTxImpl struct {
	tx *gorm.DB
}

func (t *economyTxImpl) Balance(playerID string) (int, error) {
	return balance(t.tx, playerID)
}

func (t *economyTxImpl) SumDepositCoinsSince(playerID string, since time.Time) (sum int, err error) {
	err = t.tx.Model(&dzMDL.CoinDeposit{}).
		Where("player_id = ? AND created_at >= ?", playerID, since).
		Select("COALESCE(SUM(coins), 0)").
		Scan(&sum).Error
	return
}

func (t *economyTxImpl) CountPurchases(playerID, itemKey string) (cnt int64, err error) {
	err = t.tx.Model(&dzMDL.StorePurchase{}).
		Where("player_id = ? AND item_key = ?", playerID, itemKey).
		Count(&cnt).Error
	return
}

func (t *economyTxImpl) CreateTransaction(tr *myMDL.Transaction) error {
	return t.tx.Create(tr).Error
}

func (t *economyTxImpl) CreateDeposit(row *dzMDL.CoinDeposit) error {
	return mapDuplicate(t.tx.Create(row).Error)
}

func (t *economyTxImpl) CreatePurchase(row *dzMDL.StorePurchase) error {
	return mapDuplicate(t.tx.Create(row).Error)
}
func mapDuplicate(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(msg, "Duplicate entry") || strings.Contains(msg, "1062") {
		return myERR.ErrDuckzDuplicate
	}
	return err
}

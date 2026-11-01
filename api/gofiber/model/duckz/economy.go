package duckz

import (
	"database/sql"
	"time"

	myMDL "gofiber/model"
)

// Platform values must fit transactions.platform varchar(24).
const (
	TransactionPlatformGoldDeposit   = "DUCKZ_GOLD_DEPOSIT"
	TransactionPlatformStorePurchase = "DUCKZ_STORE_PURCHASE"
	TransactionPlatformAdminAdjust   = "DUCKZ_ADMIN_ADJUST"
	TransactionPlatformRedemption    = "DUCKZ_REDEMPTION"
	TransactionPlatformCreatorGrant  = "DUCKZ_CREATOR_GRANT"
)

// NewCoinGrant is a credit-only transactions row for coins granted straight to the wallet
// (redemption codes, content creator rewards). ref is stored in platform_account, e.g. "code:12".
func NewCoinGrant(playerID string, coins int, platform, ref string) *myMDL.Transaction {
	return &myMDL.Transaction{
		TransactionCreation: myMDL.TransactionCreation{
			PlayerID:        playerID,
			Credit:          coins,
			Platform:        sql.NullString{String: platform, Valid: true},
			PlatformAccount: sql.NullString{String: ref, Valid: ref != ""},
		},
	}
}

// CoinDeposit is one gold-bar-to-coin conversion; RequestID is the client idempotency key.
type CoinDeposit struct {
	ID            uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	PlayerID      string    `json:"player_id" gorm:"column:player_id;size:128;not null"`
	RequestID     string    `json:"request_id" gorm:"column:request_id;size:64;not null"`
	Bars          int       `json:"bars" gorm:"column:bars;not null"`
	Coins         int       `json:"coins" gorm:"column:coins;not null"`
	TransactionID *uint     `json:"transaction_id,omitempty" gorm:"column:transaction_id"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;not null"`
}

func (CoinDeposit) TableName() string {
	return "duckz_coin_deposits"
}

// StorePurchase records a coin store purchase; RequestID is the client idempotency key.
type StorePurchase struct {
	ID            uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	PlayerID      string    `json:"player_id" gorm:"column:player_id;size:128;not null"`
	ItemKey       string    `json:"item_key" gorm:"column:item_key;size:128;not null"`
	Price         int       `json:"price" gorm:"column:price;not null"`
	TransactionID *uint     `json:"transaction_id,omitempty" gorm:"column:transaction_id"`
	RequestID     string    `json:"request_id" gorm:"column:request_id;size:64;not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;not null"`
}

func (StorePurchase) TableName() string {
	return "duckz_store_purchases"
}

// Entitlement is a durable IAP owned forever; unique on (PlayerID, Sku) and PlatformTxnID.
type Entitlement struct {
	ID            uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	PlayerID      string    `json:"player_id" gorm:"column:player_id;size:128;not null"`
	Sku           string    `json:"sku" gorm:"column:sku;size:128;not null"`
	Platform      string    `json:"platform" gorm:"column:platform;size:24;not null;default:''"`
	PlatformTxnID string    `json:"platform_txn_id" gorm:"column:platform_txn_id;size:512;not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;not null"`
}

func (Entitlement) TableName() string {
	return "duckz_entitlements"
}

// Loadout is the avatar a player has equipped. Empty SkinKey is the default duck.
type Loadout struct {
	PlayerID  string    `json:"player_id" gorm:"column:player_id;size:128;primaryKey"`
	SkinKey   string    `json:"skin_key" gorm:"column:skin_key;size:128;not null;default:''"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;not null"`
}

func (Loadout) TableName() string {
	return "duckz_loadouts"
}

// AdminTransaction is one `transactions` row as shown on the Duckz admin player page.
type AdminTransaction struct {
	ID              uint      `json:"transaction_id"`
	Platform        string    `json:"platform"`
	PlatformAccount string    `json:"platform_account"`
	PlatformTxnID   string    `json:"platform_txn_id"`
	Debit           int       `json:"transaction_debit"`
	Credit          int       `json:"transaction_credit"`
	IsDeleted       bool      `json:"transaction_is_deleted"`
	Date            time.Time `json:"transaction_date"`
}

// CoinAdjustment is a DUCKZ_ADMIN_ADJUST row from `transactions`: a grant (positive Amount) or removal (negative).
type CoinAdjustment struct {
	ID        uint      `json:"id"`
	AdminID   string    `json:"admin_id"`
	Amount    int       `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

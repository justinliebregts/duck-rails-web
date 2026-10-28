package model

import (
	"database/sql"
	"errors"
	myDB "gofiber/database/mysql"
	myERR "gofiber/errors"
	"time"

	"gorm.io/gorm"
)

var ErrGiveGemPlayerIDRequired = errors.New("player_id is required")

//type Transactions []Transaction
//
//func (t Transactions) CreditSum() int {
//	sum := 0
//	for _, t := range t {
//		sum += t.Credit
//	}
//	return sum
//}
//
//func (t Transactions) DebitSum() int {
//	sum := 0
//	for _, t := range t {
//		sum += t.Debit
//	}
//	return sum
//}

type Transaction struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`
	TransactionCreation
	Datetime time.Time `json:"date" gorm:"column:date;autoCreateTime"`
}

func (t *Transaction) BeforeSave(tx *gorm.DB) (err error) {
	if t.Datetime.IsZero() {
		t.Datetime = CurrentTimestamp()
	}
	return
}

type TransactionCreation struct {
	PlayerID        string         `json:"player_id"`
	ItemID          sql.NullInt32  `json:"item_id"`
	Debit           int            `json:"debit" gorm:"default:0" example:"0"`
	Credit          int            `json:"credit" gorm:"default:0" example:"0"`
	Platform        sql.NullString `json:"platform"`
	PlatformAccount sql.NullString `json:"platform_account"`
	ParentID        sql.NullInt32  `json:"parent_id"`
	PlatformTxnID   sql.NullString `json:"platform_txn_id"`
	PlatformTxnTime sql.NullString `json:"platform_txn_time"`
	IsDeleted       bool           `json:"is_deleted"`
	IsBundle        bool           `json:"is_bundle" gorm:"-"`
}

// DeleteTransactionAdmin was deleteTransactionAdmin($transaction_id, $is_deleted)
func DeleteTransactionAdmin(transactionID uint, isDeleted bool) (err error) {
	//$query = "UPDATE transactions SET is_deleted=:is_deleted WHERE id=:transaction_id";
	result := myDB.DBConn.Model(&Transaction{}).Where("id = ?", transactionID)
	return result.Update("is_deleted", isDeleted).Error
}

// DeleteAllTransactionsAdmin was deleteAllTransactionsAdmin($player_id){
func DeleteAllTransactionsAdmin(playerID string) error {
	//$query = "UPDATE transactions SET is_deleted = 1 WHERE player_id=:player_id";
	result := myDB.DBConn.Model(&Transaction{}).Where("player_id = ?", playerID)
	return result.Updates(Transaction{TransactionCreation: TransactionCreation{IsDeleted: true}}).Error
}

func OwnItem(playerID string, itemID uint) (count int64, err error) {
	result := myDB.DBConn.Model(&Transaction{}).Where("player_id = ? AND item_id = ? AND is_deleted = false", playerID, itemID)
	if result.Count(&count); result.Error != nil {
		err = result.Error
	}
	return
}

func IsPlatformTxnRecorded(platformTxnID string) (recorded bool, err error) {
	var count int64
	result := myDB.DBConn.Model(&Transaction{}).Where("platform_txn_id = ? AND is_deleted = false", platformTxnID).Count(&count)
	if result.Error != nil {
		err = result.Error
		return
	}
	recorded = count > 0
	return
}

type GetItemsResponse struct {
	TransactionID  sql.NullInt64  `json:"transaction_id"`
	ItemID         uint           `json:"item_id"`
	ItemName       string         `json:"item_name"`
	ItemType       int            `json:"item_type"`
	WeaponType     sql.NullInt16  `json:"weapon_type"`
	AssetID        sql.NullString `json:"asset_id"`
	SkinStyleID    int            `json:"skin_style_id"`
	IsDefault      bool           `json:"is_default"`
	CanOwnMultiple bool           `json:"can_own_multiple"`
}

// GetItems returns default avatar and weapon skins plus skins this player owns.
func GetItems(playerID string, devOwned bool) (rows []GetItemsResponse, err error) {
	devClause := ""
	args := []interface{}{ItemTypeWeapon, ItemTypeAvatar}
	if devOwned {
		devClause = " OR (dev_owned = 1 AND type IN (?, ?))"
		args = append(args, ItemTypeWeapon, ItemTypeAvatar)
	}
	args = append(args, playerID, ItemTypeWeapon, ItemTypeAvatar)

	query := `
SELECT NULL AS transaction_id, id AS item_id, name AS item_name, type AS item_type,
       weapon_type, asset_id, skin_style_id, is_default, can_own_multiple
FROM items
WHERE (is_default = 1 AND type IN (?, ?))` + devClause + `
UNION
SELECT transactions.id AS transaction_id, items.id AS item_id, items.name AS item_name, items.type AS item_type,
       items.weapon_type, items.asset_id, items.skin_style_id, items.is_default, items.can_own_multiple
FROM transactions
JOIN items ON items.id = transactions.item_id
WHERE transactions.player_id = ? AND transactions.is_deleted = 0 AND items.type IN (?, ?)`

	result := myDB.DBConn.Raw(query, args...).Scan(&rows)
	if result.Error != nil {
		err = result.Error
		return
	}
	if rows == nil {
		rows = []GetItemsResponse{}
	}
	return
}

type GetItemsAdminResponse struct {
	TransactionID        uint           `json:"transaction_id"`
	ItemID               uint           `json:"item_id"`
	ItemName             string         `json:"item_name"`
	SkinName             string         `json:"skin_name"`
	ItemType             string         `json:"item_type"`
	WeaponType           string         `json:"weapon_type"`
	AssetID              sql.NullString `json:"asset_id"`
	Platform             sql.NullString `json:"platform"`
	PlatformTxnID        sql.NullString `json:"platform_txn_id"`
	TransactionDate      time.Time      `json:"transaction_date"`
	TransactionIsDeleted int            `json:"transaction_is_deleted"`
}

func GetItemsAdmin(playerID string) (rows []GetItemsAdminResponse, err error) {
	sel := "transactions.id as transaction_id, items.id as item_id, items.name as item_name, skin_styles.name as skin_name, " +
		"item_types.type as item_type, weapons.name as weapon_type, items.asset_id as asset_id, " +
		"transactions.platform as platform, transactions.platform_txn_id as platform_txn_id, " +
		"transactions.date as transaction_date, transactions.is_deleted as transaction_is_deleted"
	result := myDB.DBConn.Model(&Transaction{}).Where("player_id = ?", playerID).Order("transactions.id DESC")
	result.Select(sel)
	result.Joins("LEFT JOIN items ON items.id = transactions.item_id")
	result.Joins("LEFT JOIN item_types ON item_types.id = items.type")
	result.Joins("LEFT JOIN skin_styles ON skin_styles.id = items.skin_style_id")
	result.Joins("LEFT JOIN weapons ON weapons.id = items.weapon_type")
	if err = result.Scan(&rows).Error; err != nil {
		return
	}
	if rows == nil {
		rows = []GetItemsAdminResponse{}
	}
	return
}

// GiveItemAdmin grants an avatar or weapon skin. It does not change a coin balance.
func GiveItemAdmin(playerID string, itemID uint) (t Transaction, err error) {
	var item Item
	if item, err = GetItem(itemID); err != nil {
		return
	}
	if !item.IsSkin() {
		err = myERR.ErrItemTypeNotAllowed
		return
	}

	t.TransactionCreation = TransactionCreation{
		PlayerID: playerID,
		ItemID:   sql.NullInt32{Valid: true, Int32: int32(itemID)},
		Debit:    0,
		Credit:   0,
		Platform: sql.NullString{String: "ADMIN", Valid: true},
	}
	err = myDB.DBConn.Create(&t).Error
	return
}

// RecordPurchase inserts an idempotent platform purchase for an avatar or weapon skin.
// Coin balances are not checked or changed.
func RecordPurchase(playerID string, itemID uint, platform, platformAccount, platformTxnID, platformTxnTime string) error {
	item, err := GetItem(itemID)
	if err != nil {
		return err
	}
	if !item.IsSkin() {
		return myERR.ErrItemTypeNotAllowed
	}

	if item.CanOwnMultiple == 0 {
		var count int64
		if count, err = OwnItem(playerID, itemID); err != nil {
			return err
		}
		if count > 0 {
			return myERR.ErrCannotOwnMultiple
		}
	}

	row := Transaction{
		TransactionCreation: TransactionCreation{
			PlayerID:        playerID,
			ItemID:          sql.NullInt32{Valid: true, Int32: int32(itemID)},
			Debit:           0,
			Credit:          0,
			Platform:        sql.NullString{String: platform, Valid: platform != ""},
			PlatformAccount: sql.NullString{String: platformAccount, Valid: platformAccount != ""},
			PlatformTxnID:   sql.NullString{String: platformTxnID, Valid: platformTxnID != ""},
			PlatformTxnTime: sql.NullString{String: platformTxnTime, Valid: platformTxnTime != ""},
		},
	}
	return myDB.DBConn.Create(&row).Error
}

package model

import (
	"database/sql"

	myDB "gofiber/database/mysql"
	myERR "gofiber/errors"
)

const (
	ItemTypeWeapon = 2
	ItemTypeAvatar = 3
)

type Item struct {
	ID             uint           `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name           string         `gorm:"column:name" json:"name"`
	RarityID       int            `gorm:"column:rarity_id" json:"rarity_id"`
	WeaponType     sql.NullInt16  `gorm:"column:weapon_type" json:"weapon_type"`
	AssetID        sql.NullString `gorm:"column:asset_id;type:varchar(64)" json:"asset_id"`
	SkinStyleID    int            `gorm:"column:skin_style_id" json:"skin_style_id"`
	Price          int            `gorm:"column:price" json:"price"`
	Type           int            `gorm:"column:type" json:"type"`
	IsDefault      int            `gorm:"column:is_default;default:0" json:"is_default"`
	DevOwned       int            `gorm:"column:dev_owned;default:0" json:"dev_owned"`
	CanOwnMultiple int            `gorm:"column:can_own_multiple;default:0" json:"can_own_multiple"`
	GRReady        int            `gorm:"column:gr_ready;default:0" json:"gr_ready"`
}

func (Item) TableName() string {
	return "items"
}

func (i Item) IsSkin() bool {
	return i.Type == ItemTypeWeapon || i.Type == ItemTypeAvatar
}

func (i *Item) Update() error {
	var weaponType interface{}
	if i.WeaponType.Valid {
		weaponType = i.WeaponType.Int16
	}
	var assetID interface{}
	if i.AssetID.Valid {
		assetID = i.AssetID.String
	}
	result := myDB.DBConn.Model(&Item{}).Where("id = ?", i.ID).Updates(map[string]interface{}{
		"name":             i.Name,
		"weapon_type":      weaponType,
		"asset_id":         assetID,
		"skin_style_id":    i.SkinStyleID,
		"is_default":       i.IsDefault,
		"dev_owned":        i.DevOwned,
		"can_own_multiple": i.CanOwnMultiple,
		"gr_ready":         i.GRReady,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return myERR.NoRowAffected
	}
	return nil
}

func GetItem(itemID uint) (item Item, err error) {
	result := myDB.DBConn.First(&item, itemID)
	if result.Error != nil {
		err = result.Error
	} else if result.RowsAffected == 0 {
		err = myERR.NoItemFound
	}
	return
}

func CreateItem(item Item) (inserted Item, err error) {
	if result := myDB.DBConn.Create(&item); result.Error != nil {
		err = result.Error
	} else {
		inserted = item
	}
	return
}

type AllItemsResponse struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	WeaponName     string `json:"weapon_name"`
	WeaponType     int    `json:"weapon_type"`
	SkinStyleName  string `json:"skin_style_name"`
	SkinStyleID    int    `json:"skin_style_id"`
	Type           string `json:"type"`
	TypeID         int    `json:"type_id"`
	AssetID        string `json:"asset_id"`
	IsDefault      int    `json:"is_default"`
	DevOwned       int    `json:"dev_owned"`
	CanOwnMultiple int    `json:"can_own_multiple"`
	GRReady        int    `json:"gr_ready"`
}

func AllItems() (rows []AllItemsResponse, err error) {
	sel := `items.id, items.name,
COALESCE(CONCAT(weapons.name, "_", items.weapon_type), "") as weapon_name,
COALESCE(items.weapon_type, 0) as weapon_type,
COALESCE(skin_styles.name, "") as skin_style_name, items.skin_style_id,
item_types.type as type, items.type as type_id, COALESCE(items.asset_id, "") as asset_id,
items.is_default, items.dev_owned, items.can_own_multiple, items.gr_ready`
	result := myDB.DBConn.Model(&Item{}).
		Select(sel).
		Joins("LEFT JOIN weapons ON weapons.id = items.weapon_type").
		Joins("LEFT JOIN item_types ON item_types.id = items.type").
		Joins("LEFT JOIN skin_styles ON skin_styles.id = items.skin_style_id").
		Where("items.type IN ?", []int{ItemTypeWeapon, ItemTypeAvatar}).
		Order("items.id DESC").
		Scan(&rows)
	if result.Error != nil {
		err = result.Error
		return
	}
	if rows == nil {
		rows = []AllItemsResponse{}
	}
	return
}

type Weapon struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
}

func (Weapon) TableName() string {
	return "weapons"
}

func ListWeapons() (rows []Weapon, err error) {
	result := myDB.DBConn.Order("id").Find(&rows)
	if result.Error != nil {
		err = result.Error
		return
	}
	if rows == nil {
		rows = []Weapon{}
	}
	return
}

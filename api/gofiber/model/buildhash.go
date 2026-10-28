package model

import (
	"encoding/json"
	myDB "gofiber/database/mysql"
	"sort"
	"time"

	"gorm.io/datatypes"
)

type BuildHash struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`
	BuildHashCore
	Datetime time.Time `json:"date" gorm:"column:date"`
}

func (b *BuildHash) TableName() string {
	return "build_hash"
}

type BuildHashCore struct {
	Version string         `json:"version" gorm:"not null"`
	Hash    string         `json:"hash" gorm:"not null"`
	ObbHash datatypes.JSON `json:"obb_hash" gorm:"type:json"`
}

func (b *BuildHashCore) FillObbHash(arr []string) {
	if arr == nil || len(arr) == 0 {
		b.ObbHash = datatypes.JSON([]byte(`[]`)) // Default to empty JSON array
		return
	}

	// Sort the slice if you need it sorted
	sort.Strings(arr)

	// Convert the []string to a JSON string
	jsonData, _ := json.Marshal(arr)
	b.ObbHash = jsonData
}

func NewBuildHashCore(v *ValidateBuildHashForm) BuildHashCore {
	//return model.BuildHashCore{Version: v.Version, Hash: v.Hash}

	res := BuildHashCore{Version: v.Version, Hash: v.Hash}
	res.FillObbHash(v.ObbHash)
	return res
}

func CreateHashValid(vh BuildHashCore) error {
	result := myDB.DBConn.
		Create(&BuildHash{BuildHashCore: vh, Datetime: CurrentTimestamp()})

	return result.Error
}

//func CheckHashValid(vh BuildHashCore) (valid bool, err error) {
//	var cnt int64
//	result := myDB.DBConn.
//		Model(&BuildHash{}).
//		Where(&BuildHash{BuildHashCore: vh})
//
//	//// If ObbHash is not empty, add the JSON condition
//	//if vh.ObbHash.String() != "[]" {
//	//	result = result.Where("JSON_UNQUOTE(JSON_EXTRACT(obb_hash, '$[0]')) = ?", vh.ObbHash)
//	//}
//
//	result.Count(&cnt)
//
//	if err = result.Error; err == nil {
//		valid = cnt > 0
//	}
//
//	return
//}

func RetrieveAllByVersionHash(vh BuildHashCore) (obHsh []string, err error) {

	var obbHashes []datatypes.JSON
	if err = myDB.DBConn.
		Model(&BuildHash{}).
		Where("version = ? AND hash = ?", vh.Version, vh.Hash).
		Pluck("obb_hash", &obbHashes).Error; err != nil {
		return
	}

	obHsh = make([]string, 0, len(obbHashes))
	for _, obHa := range obbHashes {
		obHsh = append(obHsh, string(obHa))
	}

	return
}

func IsHashObbHashVersionMatch(bhc BuildHashCore) (hashVersionMatched bool, obbhashHashVersionMatched bool) {
	//log.Printf("bhc: %#v \n %#v", bhc, bhc.ObbHash)
	obbHashFromDB, err := RetrieveAllByVersionHash(bhc)
	if err != nil || len(obbHashFromDB) == 0 {
		return
	}

	hashVersionMatched = true
	bhcObbString := bhc.ObbHash.String()
	//log.Printf("bhcObbString len %v:%v", len(bhcObbString), bhcObbString)
	for _, dbHash := range obbHashFromDB {
		//log.Printf("len %v:%v", len(dbHash), dbHash)
		if bhcObbString == dbHash {
			obbhashHashVersionMatched = true
			return
		}
	}

	return
}

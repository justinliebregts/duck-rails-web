package model

import (
	myDB "gofiber/database/mysql"
	myERR "gofiber/errors"
)

type SkinStyle struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name"`
}

func GetAllSkinStyles() (ss []SkinStyle, err error) {
	// An empty collection is a successful result and must encode as [], not null.
	ss = make([]SkinStyle, 0)
	err = myDB.DBConn.Find(&ss).Error
	return
}

// CreateSkinStyle was Skin.create($data)
func CreateSkinStyle(id uint, name string) (ss SkinStyle, err error) {
	// id is optional. When set, it must match the C# skin style enum.
	ss = SkinStyle{ID: id, Name: name}
	result := myDB.DBConn.Save(&ss)
	err = result.Error
	if err != nil {
		//do nothing
	} else if result.RowsAffected == 0 {
		err = myERR.NoRowAffected
	}

	return
}

// UpdateSkinStyle was Skin.update($data)
func UpdateSkinStyle(id uint, name string) error {
	// $query = "UPDATE  skin_styles SET  name=:name WHERE id = :id";

	if result := myDB.DBConn.Model(&SkinStyle{}).Where("id = ?", id).Update("name", name); result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return myERR.NoRowAffected
	}

	return nil
}

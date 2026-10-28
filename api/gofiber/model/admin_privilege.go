package model

import (
	"database/sql"
	myDB "gofiber/database/mysql"
	myERR "gofiber/errors"
	"time"
)

type AdminPrivilege struct {
	//gorm.Model
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
	PlayerID  string
}

func (a *AdminPrivilege) TableName() string {
	return "admin_privileges"
}

func (a *AdminPrivilege) Save() error {
	result := myDB.DBConn.Save(a)
	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return myERR.NoRowAffected
	}

	return nil
}

func RetrieveAdminPrivilege(playerID string) (adminPrivilege AdminPrivilege, err error) {
	adminPrivilege.PlayerID = playerID
	result := myDB.DBConn.Model(&AdminPrivilege{}).Where(&adminPrivilege).Limit(1)
	if err = result.Scan(&adminPrivilege).Error; err != nil {
		// do nothing for err != nil
	} else if result.RowsAffected == 0 {
		err = myERR.NoRowFound
	}

	return
}

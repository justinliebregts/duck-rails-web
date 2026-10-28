package model

import (
	myDB "gofiber/database/mysql"
	myERR "gofiber/errors"
)

// GetAllPlayers deprecated, should use playerREPO
func GetAllPlayers() (players []Player, err error) {
	result := myDB.DBConn.Find(&players)
	err = result.Error
	return
}

// RetrieveFromDB deprecated, should use playerREPO
func RetrieveFromDB(playerID string) (player Player, err error) {
	player.ID = playerID
	result := myDB.DBConn.First(&player)
	err = result.Error
	return
}

// SearchBySessionBase deprecated, should use playerREPO
func SearchBySessionBase(sb SessionBase) (player Player, err error) {
	//err = myDB.DBConn.Where(&Player{ID: sb.ComposedID(), Platform: sb.Platform}).First(&player).Error
	pp := &Player{ID: sb.ComposedID(), Platform: sb.Platform}
	res := myDB.DBConn.Model(pp).Where(pp).Limit(1).Scan(&player)
	if res.Error != nil {
		err = res.Error
	} else if res.RowsAffected < 1 {
		err = myERR.NoMatch
	}
	return
}

// SearchByID deprecated, should use playerREPO
func SearchByID(playerID string) (players []Player, err error) {
	//$query = "SELECT * FROM players WHERE id = :id";
	result := myDB.DBConn.Where(&Player{ID: playerID}).Find(&players)
	err = result.Error
	return
}

// SearchByName deprecated, should use playerREPO
func SearchByName(playerName string) (players []Player, err error) {
	//$query = "SELECT * FROM players WHERE name LIKE :name";
	result := myDB.DBConn.Where("name LIKE ?", "%"+playerName+"%").Find(&players)
	err = result.Error
	return
}

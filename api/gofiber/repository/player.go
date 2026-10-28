package repository

import (
	myMDL "gofiber/model"

	"gorm.io/gorm"
)

type PlayerRepository interface {
	RetrieveFromDB(playerUUID string) (*myMDL.Player, error)
	SaveToDB(playerPtr *myMDL.Player) error

	ListAdminUser() (players []myMDL.Player, err error)
}

func NewPlayerRepository(c *gorm.DB) PlayerRepository {
	return &playerRepositoryImpl{dbc: c}
}

type playerRepositoryImpl struct {
	dbc *gorm.DB
}

func (p *playerRepositoryImpl) RetrieveFromDB(playerID string) (playerPtr *myMDL.Player, err error) {
	player := myMDL.Player{ID: playerID}

	// Using the struct with ID set acts as the search criteria
	if err = p.dbc.First(&player).Error; err != nil {
		return
	}
	playerPtr = &player
	return
}

func (p *playerRepositoryImpl) SaveToDB(playerPtr *myMDL.Player) error {
	return p.dbc.Save(playerPtr).Error
}

func (p *playerRepositoryImpl) GetPlayerBalance(uuid myMDL.SessionBase) (balance int, err error) {
	result := p.dbc.
		Model(&myMDL.Transaction{}).
		Where("player_id= ? AND is_deleted=false", uuid.ComposedID()).
		Select("COALESCE(SUM(credit) - SUM(debit),0)")
	err = result.Scan(&balance).Error
	return
}

func (s *playerRepositoryImpl) ListAdminUser() (players []myMDL.Player, err error) {
	err = s.dbc.
		Model(&myMDL.Player{}).
		//Select("players.*").
		Joins("INNER JOIN admin_privileges ON players.id = admin_privileges.player_id").
		Where("admin_privileges.deleted_at IS NULL"). //Where(myMDL.AdminPrivilege{DeletedAt: sql.NullTime{Valid: false}}).
		Find(&players).Error
	return
}

package repository

import (
	"fmt"
	myDOM "gofiber/model"

	"gorm.io/gorm"
)

type DBPatchRepository interface {
	TransactionRenameUser(from myDOM.SessionBase, to myDOM.SessionBase) error
}

func NewDBPatchRepository(c *gorm.DB) DBPatchRepository {
	return &dbPatchRepositoryImpl{dbc: c}
}

type dbPatchRepositoryImpl struct {
	dbc *gorm.DB
}

func (f *dbPatchRepositoryImpl) TransactionRenameUser(from myDOM.SessionBase, to myDOM.SessionBase) error {
	return f.dbc.Transaction(func(tx *gorm.DB) error {

		fromUID := from.ComposedID()
		toUID := to.ComposedID()

		updates := []struct {
			Table       string
			Column      string
			OriginalVal string
			NewVal      string
		}{
			{"reports", "reported_id", fromUID, toUID},
			{"reports", "reporter_id", fromUID, toUID},
			{"service_bans", "target_id", fromUID, toUID},
			{"transactions", "player_id", fromUID, toUID},
		}

		for _, update := range updates {
			query := tx.Table(update.Table).
				Where(fmt.Sprintf("%s = ?", update.Column), update.OriginalVal).
				Update(update.Column, update.NewVal)

			if err := query.Error; err != nil {
				return err
			}
		}

		return nil
	})
}

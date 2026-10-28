package repository

import (
	"gofiber/model"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"regexp"
	"testing"
)

// The reduced schema must migrate retained records without querying legacy tables.
func TestRenameUserWithReducedSchema(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: conn, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	from := model.NewSessionBase("OCULUS", "old-player")
	to := model.NewSessionBase("OCULUS", "new-player")
	mock.ExpectBegin()
	for _, pair := range [][2]string{{"reports", "reported_id"}, {"reports", "reporter_id"}, {"service_bans", "target_id"}, {"transactions", "player_id"}} {
		query := "UPDATE `" + pair[0] + "` SET `" + pair[1] + "`=? WHERE " + pair[1] + " = ?"
		mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(to.ComposedID(), from.ComposedID()).WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()
	if err := NewDBPatchRepository(db).TransactionRenameUser(from, to); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

package model

import (
	myDB "gofiber/database/mysql"
	"testing"
)

func TestIPBlacklistTablesMigrator(t *testing.T) {
	t.Skip("Skipping the AutoMigrate part of the test.")
	myDB.ConnectDb()
	if myDB.DBConn == nil {
		t.Fatal("DBConn is nil after ConnectDb()")
	}
	t.Run("AutoMigrate", func(t *testing.T) {

		//Drop table if exists (will ignore or delete foreign key constraints when dropping)
		err := myDB.DBConn.Migrator().DropTable(&IPBlacklist{})
		if err != nil {
			t.Fatalf("DropTable failed: %v", err)
		}

		err = myDB.DBConn.AutoMigrate(&IPBlacklist{})
		if err != nil {
			t.Fatalf("AutoMigrate failed: %v", err)
		}
	})
}

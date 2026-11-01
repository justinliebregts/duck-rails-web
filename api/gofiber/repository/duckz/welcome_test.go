package duckz

import (
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestWelcomeGrantOnlyOnceIncludingEmptyWallet(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: conn, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	repo := NewEconomyRepository(db)
	for _, tc := range []struct {
		receipt int
		coins   int
	}{{0, 100}, {1, 100}, {1, 0}} {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT `id` FROM `players`.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("test-player"))
		mock.ExpectQuery("SELECT count.*FROM `rails_wallet_grants`").WithArgs("test-player").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tc.receipt))
		if tc.receipt == 0 {
			mock.ExpectExec("INSERT INTO rails_wallet_grants").WithArgs("test-player", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec("INSERT INTO `transactions`").WillReturnResult(sqlmock.NewResult(1, 1))
		}
		mock.ExpectQuery("SELECT COALESCE.*FROM `transactions`").WithArgs("test-player").WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(tc.coins))
		mock.ExpectCommit()
		coins, err := repo.Balance("test-player")
		if err != nil || coins != tc.coins {
			t.Fatalf("balance %d, err %v", coins, err)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWelcomeGrantRollsBackOnLedgerFailure(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: conn, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT `id` FROM `players`.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("test-player"))
	mock.ExpectQuery("SELECT count.*FROM `rails_wallet_grants`").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("INSERT INTO rails_wallet_grants").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `transactions`").WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()
	if _, err := NewEconomyRepository(db).Balance("test-player"); err == nil {
		t.Fatal("ledger failure must not consume welcome grant")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

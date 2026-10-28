package model

import (
	myDB "gofiber/database/mysql"
	"reflect"
	"strconv"
	"testing"
	"time"

	"gorm.io/datatypes"
)

var (
	sampleBuildHash  = BuildHash{BuildHashCore: BuildHashCore{Hash: "h123", Version: "1.2.3"}, Datetime: CurrentTimestamp()}
	sampleBuildHash2 = BuildHash{BuildHashCore: BuildHashCore{Hash: "h456", Version: "2.4.6", ObbHash: datatypes.JSON([]byte(`["t123"]`))}, Datetime: CurrentTimestamp()}
	sampleBuildHash3 = BuildHash{BuildHashCore: BuildHashCore{Hash: "h456", Version: "2.4.8", ObbHash: datatypes.JSON([]byte(`["t123,t456"]`))}, Datetime: CurrentTimestamp()}
)

func TestBuildHashTablesMigrator(t *testing.T) {
	t.Skip("Skipping the AutoMigrate part of the test.")
	myDB.ConnectDb()

	t.Run("AutoMigrate", func(t *testing.T) {

		// Drop table if exists (will ignore or delete foreign key constraints when dropping)
		myDB.DBConn.Migrator().DropTable(
			&BuildHash{},
		)

		err := myDB.DBConn.AutoMigrate(
			&BuildHash{},
		)
		if err != nil {
			t.Fatalf("AutoMigrate failed: %v", err)
		}
	})
}

func TestValidateBuildHashForm_CheckClientVersion(t *testing.T) {

	type fields struct {
		Version string
		Hash    string
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{"t0 only allowed x.y.z", fields{"1.2.3.4", "hash"}, true},
		{"t0-2 only allowed x.y.z", fields{"1.2.3.4.5", "hash"}, true},
		{"t1 parsed as 1.0.0", fields{"1", "hash"}, false},
		{"t1-2 parsed as 0.10.0", fields{"0.10", "hash"}, false},
		{"t2 version too low", fields{"0.8.241", "hash"}, true},
		{"t3 version too low for 0.9.1", fields{"0.9.0", "hash"}, true},
		{"t3 version too high for 1.4", fields{"1.5.0", "hash"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &ValidateBuildHashForm{
				Version: tt.fields.Version,
				Hash:    tt.fields.Hash,
			}
			if err := v.CheckClientVersion(); (err != nil) != tt.wantErr {
				t.Errorf("CheckClientVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateHashValid(t *testing.T) {

	myDB.ConnectDb()

	sampleBuildHash.FillObbHash(nil)
	sampleBuildHash2.FillObbHash([]string{"t1234t4567"})
	sampleBuildHash3.FillObbHash([]string{"t1234", "t4567"})

	tests := []struct {
		name    string
		args    BuildHashCore
		wantErr bool
	}{
		// TODO: Add test cases.
		{"t1", sampleBuildHash.BuildHashCore, false},
		{"t2", sampleBuildHash2.BuildHashCore, false},
		{"t3", sampleBuildHash3.BuildHashCore, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateHashValid(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("CreateHashValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//func TestCheckHashValid(t *testing.T) {
//
//	myDB.ConnectDb()
//
//	sampleBuildHash.FillObbHash(nil)
//	sampleBuildHash2.FillObbHash([]string{"t1234t4567"})
//	sampleBuildHash3.FillObbHash([]string{"t1234", "t4567"})
//
//	sampleBuildHash4 := sampleBuildHash3
//	sampleBuildHash4.FillObbHash([]string{"t7123", "t45679"})
//
//	tests := []struct {
//		name      string
//		args      BuildHashCore
//		wantValid bool
//		wantErr   bool
//	}{
//		// TODO: Add test cases.
//		{"t1", sampleBuildHash.BuildHashCore, true, false},
//		{"t2", BuildHashCore{Hash: "h1234", Version: "1.2.3"}, false, false},
//
//		{"t3", sampleBuildHash2.BuildHashCore, true, false},
//		{"t4", sampleBuildHash3.BuildHashCore, true, false},
//		{"t5", sampleBuildHash4.BuildHashCore, false, false},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			gotValid, err := CheckHashValid(tt.args)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("CheckHashValid() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if gotValid != tt.wantValid {
//				t.Errorf("CheckHashValid() gotValid = %v, want %v", gotValid, tt.wantValid)
//			}
//		})
//	}
//}

func TestRetrieveAllByVersionHash(t *testing.T) {

	myDB.ConnectDb()

	seedVersion := "test-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	withObb := BuildHashCore{Version: seedVersion, Hash: "h123"}
	withObb.FillObbHash([]string{"t2", "t1"})
	withoutObb := BuildHashCore{Version: seedVersion, Hash: "h123"}
	withoutObb.FillObbHash(nil)
	for _, bh := range []BuildHashCore{withObb, withoutObb} {
		if err := CreateHashValid(bh); err != nil {
			t.Fatalf("CreateHashValid() error = %v", err)
		}
	}
	t.Cleanup(func() {
		myDB.DBConn.Where("version = ?", seedVersion).Delete(&BuildHash{})
	})

	type args struct {
		vh BuildHashCore
	}
	tests := []struct {
		name     string
		args     args
		wantHash []string
		wantErr  bool
	}{
		{"t0 empty", args{BuildHashCore{}}, []string{}, false},
		{"t0-2 no result", args{BuildHashCore{Version: seedVersion, Hash: "other"}}, []string{}, false},
		{"t1", args{BuildHashCore{Version: seedVersion, Hash: "h123"}}, []string{"[\"t1\", \"t2\"]", "[]"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHash, err := RetrieveAllByVersionHash(tt.args.vh)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveAllVersionHash() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotHash, tt.wantHash) {
				t.Errorf("RetrieveAllVersionHash() gotHash = %#v, want %#v", gotHash, tt.wantHash)
			}
		})
	}
}

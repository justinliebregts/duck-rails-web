package model

// import (
// 	"gofiber/cache"
// 	"reflect"
// 	"testing"
// 	"time"
// )

// var testVoiceRoomRecord = VoiceRoomRecord{"agora", "cn-kkkkkk", VoiceRoomRecordInCache{"t02"}}

// func TestLoadFromCache(t *testing.T) {

// 	cache.ConnectDb()

// 	tests := []struct {
// 		name    string
// 		args    string
// 		wantVrr VoiceRoomRecord
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 		{"t0", "NoSuchKey", VoiceRoomRecord{".", "NoSuchKey", VoiceRoomRecordInCache{}}, true},
// 		{"t1", testVoiceRoomRecord.MainKey(), testVoiceRoomRecord, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			gotVrr, err := LoadFromCache(tt.args)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("LoadFromCache() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(gotVrr, tt.wantVrr) {
// 				t.Errorf("LoadFromCache() gotVrr = %v, want %v", gotVrr, tt.wantVrr)
// 			}
// 		})
// 	}
// }

// func TestVoiceRoomRecord_SaveToCache(t *testing.T) {

// 	cache.ConnectDb()

// 	tests := []struct {
// 		name        string
// 		fields      VoiceRoomRecord
// 		wantSuccess bool
// 		wantErr     bool
// 	}{
// 		// TODO: Add test cases.
// 		{"t01", VoiceRoomRecord{"photon", "na-0000000", VoiceRoomRecordInCache{time.Now().String()}}, true, false},
// 		{"t02", testVoiceRoomRecord, true, false},
// 		{"t03", VoiceRoomRecord{testVoiceRoomRecord.Platform, testVoiceRoomRecord.RoomName, VoiceRoomRecordInCache{time.Now().String()}}, false, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			v := &VoiceRoomRecord{
// 				Platform:               tt.fields.Platform,
// 				RoomName:               tt.fields.RoomName,
// 				VoiceRoomRecordInCache: tt.fields.VoiceRoomRecordInCache,
// 			}
// 			gotSuccess, err := v.SaveToCache()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("SaveToCache() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if gotSuccess != tt.wantSuccess {
// 				t.Errorf("SaveToCache() gotSuccess = %v, want %v", gotSuccess, tt.wantSuccess)
// 			}
// 		})
// 	}
// }

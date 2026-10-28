package handler

import (
	"gofiber/model"
	"testing"
	"time"
)

func TestPlatformStatsResponse_CalculateRatio(t *testing.T) {

	tests := []struct {
		name   string
		fields PlatformStatsResponse
	}{
		// TODO: Add test cases.
		{"t1",
			PlatformStatsResponse{
				model.JsonTime{time.Date(1, 2, 3, 4, 5, 6, 7, time.UTC)},
				model.JsonTime{time.Date(1, 2, 3, 4, 5, 6, 7, time.UTC)},
				[]model.PlatformRatio{},
				[]model.PlatformRatio{},
				[]model.PlatformCount{{"p1", 2}, {"p2", 1}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("%#v %#v", tt.fields, len(tt.fields.platformCount))
			tt.fields.CalculateRatio()

			t.Logf("%#v", tt.fields)
		})
	}
}

// func TestSessionCreateForm_CreateNonceChecksToCache(t *testing.T) {

// 	myDB.ConnectDb()
// 	mySTO.ConnectDb()

// 	tests := []struct {
// 		name    string
// 		fields  myDOM.SessionCreateForm
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 		{"t0", myDOM.SessionCreateForm{SessionBase: myDOM.SessionBase{"t123", "GOOGLE"}, PlatformToken: "t456"}, false},
// 		{"t1", myDOM.SessionCreateForm{SessionBase: myDOM.SessionBase{"t234", "OCULUS"}, PlatformToken: "t567"}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.fields.CreateNonceChecksToCache(); (err != nil) != tt.wantErr {
// 				t.Errorf("CreateNonceChecksToCache() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestSessionCreateForm_ValidateWithDevicePlatform(t *testing.T) {

// 	myDB.ConnectDb()
// 	mySTO.ConnectDb()
// 	remote_auth.ShouldVerifyPlatformAccountWithRemoteAPI = true

// 	tests := []struct {
// 		name        string
// 		fields      SessionCreateForm
// 		wantErrMsgs []string
// 	}{
// 		// TODO: Add test cases.
// 		{"t0", SessionCreateForm{SessionBase: myDOM.SessionBase{"t123", "GOOGLE"}, PlatformToken: "t456"}, nil},
// 		{"t1", SessionCreateForm{SessionBase: myDOM.SessionBase{"t234", "OCULUS"}, PlatformToken: "t567"}, []string{"no found"}},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if gotErrMsgs := tt.fields.ValidateWithDevicePlatform(); !reflect.DeepEqual(gotErrMsgs, tt.wantErrMsgs) {
// 				t.Errorf("ValidateWithDevicePlatform() = %#v, want %v", gotErrMsgs, tt.wantErrMsgs)
// 			}
// 		})
// 	}
// }

// func TestSessionCreateForm_SkipForBuildPlatform(t *testing.T) {

// 	tests := []struct {
// 		name           string
// 		fields         SessionCreateForm
// 		wantShouldSkip bool
// 	}{
// 		// TODO: Add test cases.
// 		{"t0", SessionCreateForm{}, false},
// 		{"t0-2", SessionCreateForm{BuildPlatform: ""}, false},
// 		{"t1", SessionCreateForm{BuildPlatform: "pcvr"}, true},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if gotShouldSkip := tt.fields.SkipForBuildPlatform(); gotShouldSkip != tt.wantShouldSkip {
// 				t.Errorf("SkipForBuildPlatform() = %v, want %v", gotShouldSkip, tt.wantShouldSkip)
// 			}
// 		})
// 	}
// }

// func TestIsHashObbHashVersionMatch(t *testing.T) {

// 	myDB.ConnectDb()

// 	type args struct {
// 		bhc model.BuildHashCore
// 	}
// 	tests := []struct {
// 		name                          string
// 		args                          args
// 		wantHashVersionMatched        bool
// 		wantObbhashHashVersionMatched bool
// 	}{
// 		// TODO: Add test cases.
// 		{"t0 empty", args{model.BuildHashCore{}}, false, false},
// 		{"t0-2 not exist", args{model.BuildHashCore{Version: "11.2"}}, false, false},
// 		{"t0-3 not exist", args{model.BuildHashCore{Version: "11.2", Hash: "t123345"}}, false, false},

// 		{"t1 exist", args{model.BuildHashCore{Version: "1.2.3", Hash: "h123", ObbHash: []byte("[]")}}, true, true},
// 		{"t1-2 exist", args{model.BuildHashCore{Version: "1.2.3", Hash: "h123", ObbHash: []byte("[\"t1\", \"t2\"]")}}, true, true},
// 		{"t1-3 exist but reversed", args{model.BuildHashCore{Version: "1.2.3", Hash: "h123", ObbHash: []byte("[\"t2\", \"t1\"]")}}, true, false},
// 		{"t1-4 exist", args{model.BuildHashCore{Version: "10.10.4", Hash: "BFE6F2027F5EA1C2B7C519EFAF213E087F879877", ObbHash: []byte("[]")}}, true, true},
// 		{"t105 obbhash not exist", args{model.BuildHashCore{Version: "1.2.3", Hash: "h123", ObbHash: []byte("t321312")}}, true, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			gotHashVersionMatched, gotObbhashHashVersionMatched := IsHashObbHashVersionMatch(tt.args.bhc)
// 			if gotHashVersionMatched != tt.wantHashVersionMatched {
// 				t.Errorf("IsHashObbHashVersionMatch() gotHashVersionMatched = %v, want %v", gotHashVersionMatched, tt.wantHashVersionMatched)
// 			}
// 			if gotObbhashHashVersionMatched != tt.wantObbhashHashVersionMatched {
// 				t.Errorf("IsHashObbHashVersionMatch() gotObbhashHashVersionMatched = %v, want %v", gotObbhashHashVersionMatched, tt.wantObbhashHashVersionMatched)
// 			}
// 		})
// 	}
// }

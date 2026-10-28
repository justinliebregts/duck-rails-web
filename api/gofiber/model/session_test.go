package model

// import (
// 	"fmt"
// 	"github.com/gofiber/fiber/v2/utils"
// 	mySTO "gofiber/cache"
// 	myDB "gofiber/database/mysql"
// 	"reflect"
// 	"testing"
// )

// var (
// 	prebuildSess = Session{SessionBase: SessionBase{PlayerID: "tester007"}, AccessToken: utils.UUIDv4(), SessionRefresh: SessionRefresh{utils.UUIDv4()}}
// 	predefined   = Session{SessionBase: SessionBase{PlayerID: "t1"}, AccessToken: "t2", SessionRefresh: SessionRefresh{"t3"}}
// 	empty        = Session{}
// 	withPlayerID = Session{SessionBase: SessionBase{PlayerID: "123123"}}
// )

// //func TestBuildSession(t *testing.T) {
// //	tests := []struct {
// //		name    string
// //		args    Session
// //		wantErr bool
// //	}{
// //		{
// //			name:    "t000",
// //			args:    empty,
// //			wantErr: false,
// //		},
// //		{
// //			name:    "t001",
// //			args:    predefined,
// //			wantErr: false,
// //		},
// //		{
// //			name:    "t001 again",
// //			args:    predefined,
// //			wantErr: false,
// //		},
// //		{
// //			name:    "t002",
// //			args:    withPlayerID,
// //			wantErr: false,
// //		},
// //	}
// //
// //	for _, tt := range tests {
// //		t.Run(tt.name, func(t *testing.T) {
// //			got := BuildSession(tt.args.PlayerID, tt.args.AccessToken, tt.args.RefreshToken)
// //			t.Logf("%v", got)
// //		})
// //	}
// //}

// func TestSession_Validate(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		fields      Session
// 		wantErrMsgs []string
// 	}{
// 		// TODO: Add test cases.
// 		{
// 			name:   "test001",
// 			fields: Session{SessionBase: SessionBase{PlayerID: ""}, AccessToken: "", SessionRefresh: SessionRefresh{}},
// 			wantErrMsgs: []string{
// 				"Key: 'Session.SessionBase.PlayerID' Error:Field validation for 'PlayerID' failed on the 'required' tag",
// 				"Key: 'Session.SessionBase.AccessToken' Error:Field validation for 'AccessToken' failed on the 'required' tag",
// 				"Key: 'Session.SessionRefresh.RefreshToken' Error:Field validation for 'RefreshToken' failed on the 'required' tag",
// 			},
// 		},
// 		{
// 			name:   "test002",
// 			fields: Session{SessionBase: SessionBase{PlayerID: "1"}, AccessToken: "2", SessionRefresh: SessionRefresh{"3"}},
// 			wantErrMsgs: []string{
// 				"Key: 'Session.SessionBase.PlayerID' Error:Field validation for 'PlayerID' failed on the 'min' tag",
// 				"Key: 'Session.SessionBase.AccessToken' Error:Field validation for 'AccessToken' failed on the 'min' tag",
// 				"Key: 'Session.SessionRefresh.RefreshToken' Error:Field validation for 'RefreshToken' failed on the 'min' tag",
// 			},
// 		},
// 		{
// 			name:   "test003",
// 			fields: empty,
// 			wantErrMsgs: []string{
// 				"Key: 'Session.SessionBase.PlayerID' Error:Field validation for 'PlayerID' failed on the 'required' tag",
// 				"Key: 'Session.SessionBase.AccessToken' Error:Field validation for 'AccessToken' failed on the 'required' tag",
// 				"Key: 'Session.SessionRefresh.RefreshToken' Error:Field validation for 'RefreshToken' failed on the 'required' tag",
// 			},
// 		},
// 		{
// 			name:   "test004",
// 			fields: withPlayerID,
// 			wantErrMsgs: []string{
// 				"Key: 'Session.SessionBase.PlayerID' Error:Field validation for 'PlayerID' failed on the 'min' tag",
// 				"Key: 'Session.SessionBase.AccessToken' Error:Field validation for 'AccessToken' failed on the 'required' tag",
// 				"Key: 'Session.SessionRefresh.RefreshToken' Error:Field validation for 'RefreshToken' failed on the 'required' tag",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if gotErrMsgs := tt.fields.Validate(); !reflect.DeepEqual(gotErrMsgs, tt.wantErrMsgs) {
// 				t.Errorf("Validate() = %v, want %v", gotErrMsgs, tt.wantErrMsgs)
// 			}
// 		})
// 	}
// }

// func TestSession_SaveToCache(t *testing.T) {

// 	mySTO.ConnectDb()

// 	tests := []struct {
// 		name    string
// 		sess    Session
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 		{"sess01", Session{SessionBase{"", "OCULUS"}, "", SessionRefresh{}}, false},
// 		{"sess02", Session{SessionBase{"pid_123", "OCULUS"}, "at_456", SessionRefresh{"789"}}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.sess.SaveToCache(); (err != nil) != tt.wantErr {
// 				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestCreate(t *testing.T) {

// 	mySTO.ConnectDb()

// 	tests := []struct {
// 		name     string
// 		playerID string
// 		wantErr  bool
// 	}{
// 		{
// 			name:     "tester007",
// 			playerID: "",
// 			wantErr:  false,
// 		},
// 		{
// 			name:     "tester008",
// 			playerID: "tester008",
// 			wantErr:  false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			gotSes, err := CreatePlayerCache(tt.playerID, "tester")
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}

// 			if exi, err := gotSes.Exist(); err != nil || !exi {
// 				t.Errorf("Create() Key:%v NotInCache, err:%v", gotSes.CacheAccTknKey(), err)
// 			} else {
// 				//t.Logf("Create() Key:%v InCache", gotSes.buildKey())
// 			}

// 			if sesInCache, err := RetrievePlayerCache(gotSes.AccessToken); err != nil || !reflect.DeepEqual(gotSes, sesInCache) {
// 				t.Errorf("Create() sesInCache = %v, createdSes %v", sesInCache, gotSes)
// 			} else {
// 				//t.Logf("Create() sesInCache = %v", sesInCache)
// 			}
// 		})
// 	}
// }

// func TestCreate_Retrieve_DeleteSessionFromCache(t *testing.T) {
// 	mySTO.ConnectDb()

// 	tests := []struct {
// 		name    string
// 		args    string
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 		{
// 			name:    "test001",
// 			args:    "testerKey001",
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			gotSes, err := CreatePlayerCache(tt.args, "test2")
// 			if err != nil {
// 				t.Errorf("CreatePlayerCache() error = %v", err)
// 				return
// 			}

// 			//t.Logf("CreatePlayerCache() created = %v", gotSes)

// 			gotSes2, err := RetrievePlayerCache(gotSes.AccessToken)
// 			if err != nil {
// 				t.Errorf("RetrievePlayerCache() error = %v", err)
// 				return
// 			}
// 			if err != nil {
// 				t.Errorf("RetrievePlayerCache() error = %v", err)
// 				return
// 			}
// 			if !reflect.DeepEqual(gotSes2, gotSes) {
// 				t.Errorf("RetrievePlayerCache() = %v, want %v", gotSes2, gotSes)
// 				return
// 			}

// 			//t.Logf("RetrievePlayerCache() Equaled")

// 			if err = DeleteSessionFromCache(gotSes.AccessToken); err != nil {
// 				t.Errorf("DeletePlayerCache() error = %v", err)
// 				return
// 			}

// 			//Second time, will get nothing to delete
// 			if err = DeleteSessionFromCache(gotSes.AccessToken); err == nil {
// 				t.Errorf("DeletePlayerCache() NoError?")
// 				return
// 			}
// 		})
// 	}
// }

// func BenchmarkCreateDeleteSessionFromCache(b *testing.B) {

// 	mySTO.ConnectDb()

// 	for i := 0; i < b.N; i++ {
// 		gotSes, _ := CreatePlayerCache(fmt.Sprint("test_player", i), "test3")
// 		gotSes.Delete()
// 	}
// }

// func BenchmarkCreateRetrieveDeleteSessionFromCache(b *testing.B) {

// 	mySTO.ConnectDb()

// 	for i := 0; i < b.N; i++ {
// 		gotSes, _ := CreatePlayerCache(fmt.Sprint("test_player", i), "test4")
// 		gotSes2, _ := RetrievePlayerCache(gotSes.AccessToken)
// 		gotSes2.Delete()
// 	}
// }

// func TestSessionBase_IsGameBanned(t *testing.T) {

// 	myDB.ConnectDb()

// 	tests := []struct {
// 		name    string
// 		fields  SessionBase
// 		want    bool
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 		{"t1", SessionBase{PlayerID: "t123", Platform: "OCULUS"}, false, false},
// 		{"t2", SessionBase{PlayerID: "t123", Platform: "PICO"}, false, false},
// 		{"t3", SessionBase{PlayerID: "t456", Platform: "OCULUS"}, true, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {

// 			got, err := tt.fields.IsGameBanned()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("IsGameBanned() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got != tt.want {
// 				t.Errorf("IsGameBanned() got = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestSessionBase_TryGetPlayer(t *testing.T) {

// 	myDB.ConnectDb()
// 	mySTO.ConnectDb()

// 	tests := []struct {
// 		name    string
// 		fields  SessionBase
// 		want    Player
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := tt.fields.TryGetPlayerFromDB()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("TryGetPlayer() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("TryGetPlayer() got = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestSessionBase_TryGetDisplayName(t *testing.T) {

// 	myDB.ConnectDb()
// 	mySTO.ConnectDb()

// 	tests := []struct {
// 		name      string
// 		fields    SessionBase
// 		wantDname string
// 	}{
// 		// TODO: Add test cases.
// 		{"t0", FormSessionBase("tester123"), "UnKnOwN"},
// 		{"t0", FormSessionBase("5007278449310176"), "tester123"},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got, Dname := tt.fields.TryGetDisplayNameFromCache(); Dname != tt.wantDname {
// 				t.Errorf("TryGetDisplayName() = %v, %v, want %v", got, Dname, tt.wantDname)
// 			}
// 		})
// 	}
// }

// func TestSessionBase_CacheGetPlayerState(t *testing.T) {

// 	mySTO.ConnectDb()

// 	tests := []struct {
// 		name     string
// 		fields   SessionBase
// 		wantForm HeartbeatForm
// 		wantErr  bool
// 	}{
// 		// TODO: Add test cases.
// 		{"t1", SessionBase{PlayerID: "uid", Platform: "fake"}, HeartbeatForm{}, true},
// 		{"t2", SessionBase{PlayerID: "roombot001", Platform: "ROOMBOT"}, HeartbeatForm{}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {

// 			gotForm, err := tt.fields.CacheGetPlayerState()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("CacheGetPlayerState() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}

// 			t.Logf("%#v", gotForm)

// 			if !reflect.DeepEqual(gotForm, tt.wantForm) {
// 				t.Errorf("CacheGetPlayerState() gotForm = %v, want %v", gotForm, tt.wantForm)
// 			}
// 		})
// 	}
// }

// //func TestSessionBase_CacheSavePlayerState(t *testing.T) {
// //
// //	mySTO.ConnectDb()
// //
// //	tests := []struct {
// //		name              string
// //		fields            SessionBase
// //		args              HeartbeatForm
// //		wantRecordInCache bool
// //		wantErr           bool
// //	}{
// //		// TODO: Add test cases.
// //		{"t1", SessionBase{PlayerID: "uid", Platform: "fake"}, HeartbeatForm{"usw", "All", "54321", "Hub", "Normal", "Dune", "123", "False", "False"}, false, false},
// //		{"t2", SessionBase{PlayerID: "uid", Platform: "fake"}, HeartbeatForm{"eu", "All", "54321", "Hub", "Normal", "Dune", "123", "False", "False"}, true, false},
// //	}
// //	for _, tt := range tests {
// //		t.Run(tt.name, func(t *testing.T) {
// //
// //			gotRecordInCache, err := tt.fields.CacheSavePlayerState(tt.args)
// //			if (err != nil) != tt.wantErr {
// //				t.Errorf("CacheSavePlayerState() error = %v, wantErr %v", err, tt.wantErr)
// //				return
// //			}
// //			if gotRecordInCache != tt.wantRecordInCache {
// //				t.Errorf("CacheSavePlayerState() gotRecordInCache = %v, want %v", gotRecordInCache, tt.wantRecordInCache)
// //			}
// //		})
// //	}
// //}

// func TestSessionBase_IsInAnyValidBan(t *testing.T) {

// 	myDB.ConnectDb()

// 	type fields struct {
// 		PlayerID string
// 		Platform string
// 	}
// 	tests := []struct {
// 		name       string
// 		fields     fields
// 		wantBanned bool
// 		wantErr    bool
// 	}{
// 		// TODO: Add test cases.
// 		{"t0", fields{PlayerID: "tester123", Platform: "RoomBot"}, true, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			sb := &SessionBase{
// 				PlayerID: tt.fields.PlayerID,
// 				Platform: tt.fields.Platform,
// 			}
// 			gotBanned, err := sb.IsInAnyValidBan()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("IsInAnyValidBan() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if gotBanned != tt.wantBanned {
// 				t.Errorf("IsInAnyValidBan() gotBanned = %v, want %v", gotBanned, tt.wantBanned)
// 			}
// 		})
// 	}
// }

// func TestRetrievePlayerCacheUnderUUID(t *testing.T) {

// 	mySTO.ConnectDb()

// 	type args struct {
// 		uuid string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantSes Session
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 		{"t0", args{"fake-tester123"}, Session{}, true},

// 		{"t1", args{"3724467550970486"}, Session{SessionBase: FormSessionBase("3724467550970486"), AccessToken: "Xe94YkdkD9ejWG5rp42nugeHnEr"}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			gotSes, err := RetrievePlayerCacheUnderUUID(tt.args.uuid)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("RetrievePlayerCacheUnderUUID() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}

// 			t.Logf("%#v", gotSes)

// 			if !reflect.DeepEqual(gotSes, tt.wantSes) {
// 				t.Errorf("RetrievePlayerCacheUnderUUID() gotSes = %v, want %v", gotSes, tt.wantSes)
// 			}
// 		})
// 	}
// }

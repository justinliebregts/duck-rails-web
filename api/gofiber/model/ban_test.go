package model

var (
	fakeUserID = []string{"tester_player_001", "tester_player_002"}
)

//
//func TestGetAllBansAdmin(t *testing.T) {
//	myDB.ConnectDb()
//
//	t.Run("TestGetAllBansAdmin", func(t *testing.T) {
//		gotBanAdmins, err := GetAllBansAdmin()
//		if err != nil {
//			t.Errorf("GetAllBansAdmin() error = %v", err)
//			return
//		}
//		t.Logf("GetAllBansAdmin() %v", gotBanAdmins)
//	})
//}
//
//func TestGetAllActiveBan(t *testing.T) {
//
//	myDB.ConnectDb()
//
//	tests := []struct {
//		name      string
//		wantSbans []BanActive
//		wantErr   bool
//	}{
//		// TODO: Add test cases.
//		{"t1", []BanActive{}, false},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			gotSbans, err := GetAllActiveBan()
//			for _, val := range gotSbans {
//				t.Logf("%#v", val.PlayerID)
//			}
//
//			if (err != nil) != tt.wantErr {
//				t.Errorf("GetAllActiveBan() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(gotSbans, tt.wantSbans) {
//				t.Errorf("GetAllActiveBan() gotSbans = %#v, want %#v", gotSbans, tt.wantSbans)
//			}
//		})
//	}
//}
//
////func TestBanPlayerAdmin(t *testing.T) {
////
////	myDB.ConnectDb()
////
////	type args struct {
////		player_id   string
////		count       int
////		ban_type    BanType
////		isAutomated bool
////	}
////
////	tests := []struct {
////		name    string
////		args    args
////		wantErr bool
////	}{
////		{
////			name:    "Insert 001",
////			args:    args{fakeUserID[0], 3, BanType("TestBan"), true},
////			wantErr: false,
////		},
////		{
////			name:    "Insert 001 again",
////			args:    args{fakeUserID[0], 3, BanType("TestBan"), true},
////			wantErr: false,
////		},
////		{
////			name:    "Insert 002",
////			args:    args{fakeUserID[1], 3, BanType("TestBan2"), false},
////			wantErr: false,
////		},
////	}
////
////	for _, tt := range tests {
////		t.Run(tt.name, func(t *testing.T) {
////			if err := BanPlayerAdmin(tt.args.player_id, tt.args.count, tt.args.ban_type, tt.args.isAutomated); (err != nil) != tt.wantErr {
////				t.Errorf("BanPlayerAdmin() error = %v, wantErr %v", err, tt.wantErr)
////			}
////		})
////	}
////}
//
//func TestEndBanAdmin(t *testing.T) {
//	myDB.ConnectDb()
//
//	type args struct {
//		ban_id uint64
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//	}{
//		{
//			name:    "test01",
//			args:    args{ban_id: 1},
//			wantErr: false,
//		},
//		{
//			name:    "test02",
//			args:    args{ban_id: 5679},
//			wantErr: false,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if _, err := EndBanAdmin(tt.args.ban_id); (err != nil) != tt.wantErr {
//				t.Errorf("EndBanAdmin() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
////func TestGetPlayerBanInfo(t *testing.T) {
////
////	myDB.ConnectDb()
////	myDB.DBConn.Logger.LogMode(logger.LogLevel(4))
////
////	tests := []struct {
////		name      string
////		playerID  string
////		wantBinfo BanInfo
////		wantErr   bool
////	}{
////		{
////			name:     "test01",
////			playerID: fakeUserID[0],
////			wantBinfo: BanInfo{
////				MicBanCount:  0,
////				GameBanCount: 0,
////			},
////			wantErr: false,
////		},
////		{
////			name:     "test02",
////			playerID: "string",
////			wantBinfo: BanInfo{
////				MicBanCount:  1,
////				GameBanCount: 0,
////			},
////			wantErr: false,
////		},
////	}
////
////	for _, tt := range tests {
////		t.Run(tt.name, func(t *testing.T) {
////			gotBinfo, err := GetPlayerBanInfo(tt.playerID)
////			if (err != nil) != tt.wantErr {
////				t.Errorf("BanInfo() error = %v, wantErr %v", err, tt.wantErr)
////				return
////			}
////
////			t.Logf("BanInfo() pid:%v Got %v", tt.playerID, gotBinfo)
////
////			//if !reflect.DeepEqual(gotBinfo, tt.wantBinfo) {
////			//	t.Errorf("BanInfo() gotBinfo = %v, want %v", gotBinfo, tt.wantBinfo)
////			//}
////		})
////	}
////}
//
////func TestGetPlayerBans(t *testing.T) {
////
////	myDB.ConnectDb()
////
////	tests := []struct {
////		name          string
////		playerID      string
////		wantBansCount int
////		wantErr       bool
////	}{
////		// TODO: Add test cases.
////		{"test001", "not_a_real_player_id", 0, false},
////		{"test002", "8778586792155301", 2, false},
////		{"test003", "8457365674337157", 4, false},
////	}
////
////	for _, tt := range tests {
////		t.Run(tt.name, func(t *testing.T) {
////			gotBans, err := GetPlayerBans(tt.playerID, true, 1)
////			if (err != nil) != tt.wantErr {
////				t.Errorf("GetPlayerBans() error = %v, wantErr %v", err, tt.wantErr)
////				return
////			}
////			//if !reflect.DeepEqual(gotBans, tt.wantBans) {
////			//	t.Errorf("GetPlayerBans() gotBans = %v, want %v", gotBans, tt.wantBans)
////			//}
////			if len(gotBans) != tt.wantBansCount {
////				t.Errorf("GetPlayerBans() len want %v got %v", tt.wantBansCount, len(gotBans))
////			}
////		})
////	}
////}
////
////func TestGetPlayerBan(t *testing.T) {
////
////	myDB.ConnectDb()
////
////	tests := []struct {
////		name     string
////		playerID string
////		wantBan  Ban
////		wantErr  bool
////	}{
////		// TODO: Add test cases.
////		{"t1", "3724467550970486", Ban{}, false},
////	}
////	for _, tt := range tests {
////		t.Run(tt.name, func(t *testing.T) {
////			gotBan, err := GetPlayerBan(tt.playerID)
////			if (err != nil) != tt.wantErr {
////				t.Errorf("GetPlayerBan() error = %v, wantErr %v", err, tt.wantErr)
////				return
////			}
////			if !reflect.DeepEqual(gotBan, tt.wantBan) {
////				t.Errorf("GetPlayerBan() gotBan = %v, want %v", gotBan, tt.wantBan)
////			}
////		})
////	}
////}
//
////func TestGetPlayerSessionBan(t *testing.T) {
////
////	myDB.ConnectDb()
////
////	tests := []struct {
////		name     string
////		playerID string
////		wantBan  Ban
////		wantErr  bool
////	}{
////		// TODO: Add test cases.
////		{"T1", "", Ban{}, false},
////	}
////	for _, tt := range tests {
////		t.Run(tt.name, func(t *testing.T) {
////			gotBan, err := GetPlayerGameBan(tt.playerID)
////			if (err != nil) != tt.wantErr {
////				t.Errorf("GetPlayerSessionBan() error = %v, wantErr %v", err, tt.wantErr)
////				return
////			}
////			if !reflect.DeepEqual(gotBan, tt.wantBan) {
////				t.Errorf("GetPlayerSessionBan() gotBan = %v, want %v", gotBan, tt.wantBan)
////			}
////		})
////	}
////}
//
////func TestGetPlayerLastBan(t *testing.T) {
////
////	myDB.ConnectDb()
////
////	tests := []struct {
////		name    string
////		args    string
////		wantBan Ban
////		wantErr bool
////	}{
////		// TODO: Add test cases.
////		{"T1", "t123", Ban{}, false},
////		{"T2", "t456", Ban{}, false},
////	}
////	for _, tt := range tests {
////		t.Run(tt.name, func(t *testing.T) {
////			gotBan, err := GetPlayerLastBan(tt.args)
////			if (err != nil) != tt.wantErr {
////				t.Errorf("GetPlayerLastBan() error = %v, wantErr %v", err, tt.wantErr)
////				return
////			}
////			if !reflect.DeepEqual(gotBan, tt.wantBan) {
////				t.Errorf("GetPlayerLastBan() gotBan = %v, want %v", gotBan, tt.wantBan)
////			}
////
////			t.Logf("%#v", gotBan)
////		})
////	}
////}
//
////func TestIsPlayerGameBan(t *testing.T) {
////
////	myDB.ConnectDb()
////
////	tests := []struct {
////		name       string
////		playerID   string
////		wantBanned bool
////		wantErr    bool
////	}{
////		// TODO: Add test cases.
////		{"t1", "t1", true, false},
////		{"t2", "t2", false, false},
////		//{"t3", "t3", true, true}, // never happen
////	}
////	for _, tt := range tests {
////		t.Run(tt.name, func(t *testing.T) {
////			gotBanned, err := IsPlayerGameBan(tt.playerID)
////			if (err != nil) != tt.wantErr {
////				t.Errorf("IsPlayerGameBan() error = %v, wantErr %v", err, tt.wantErr)
////				return
////			}
////			if gotBanned != tt.wantBanned {
////				t.Errorf("IsPlayerGameBan() gotBanned = %v, want %v", gotBanned, tt.wantBanned)
////			}
////		})
////	}
////}

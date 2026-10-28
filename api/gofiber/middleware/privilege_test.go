package middleware

// func Test_checkAdminPrivilege(t *testing.T) {

// 	myDB.ConnectDb()

// 	tests := []struct {
// 		name         string
// 		playerUUID   string
// 		wantErr      error
// 		wantExtraLog string
// 	}{
// 		// TODO: Add test cases.
// 		{"t1", "z|t12345", nil, ""},
// 		{"t2", "z|t12345_NotExist", errors.New("record not found"), myERR.NoAdminPrivilegeFound.Error()},
// 		{"t3", "z|t12345_expired", errPrivilegeExpired, ""},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			gotErr, gotExtraLog := checkAdminPrivilege(tt.playerUUID)
// 			if !reflect.DeepEqual(gotErr, tt.wantErr) {
// 				t.Errorf("checkAdminPrivilege() gotErr = %v, want %v", gotErr, tt.wantErr)
// 			}
// 			if gotExtraLog != tt.wantExtraLog {
// 				t.Errorf("checkAdminPrivilege() gotExtraLog = %v, want %v", gotExtraLog, tt.wantExtraLog)
// 			}
// 		})
// 	}
// }

// func Test_checkSelfPrivilege(t *testing.T) {

// 	type args struct {
// 		playerUUID string
// 		paths      []string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr error
// 	}{
// 		// TODO: Add test cases.
// 		{"t1", args{"t123", nil}, myERR.ErrSelfPrivilegePathMissing},
// 		{"t1-2", args{"t123", []string{}}, myERR.ErrSelfPrivilegePathMissing},
// 		{"t2", args{"t456", []string{"api", "v99", "friend", "list", "t456", "t789"}}, myERR.ErrSelfPrivilegeNotOwner},
// 		{"t3", args{"t789", []string{"api", "v99", "friend", "list", "t789"}}, nil},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			gotErr := checkSelfPrivilege(tt.args.playerUUID, tt.args.paths)
// 			if !reflect.DeepEqual(gotErr, tt.wantErr) {
// 				t.Errorf("checkSelfPrivilege() gotErr = %v, want %v", gotErr, tt.wantErr)
// 			}
// 		})
// 	}
// }

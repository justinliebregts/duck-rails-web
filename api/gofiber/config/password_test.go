package config

// import (
// 	"gofiber/helper_func"
// )

// var (
// 	prov       = helper_func.BCryptHashPROV{}
// 	sha1prov   = helper_func.SHA1HashPROV{}
// 	sha256prov = helper_func.SHA256HashPROV{}
// )

//
//func TestHashPasswordEncoding(t *testing.T) {
//	tests := []struct {
//		name string
//		args string
//		//want    []byte
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//		{"t0", "t1", false},
//		{"t1", "fdgdfgt1", false},
//		{"t2", time.Now().String(), false},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			for _, assignedProv := range []HashProvider{&prov, &sha1prov, &sha256prov} {
//				for _, ty := range []helper_func.EncodingType{helper_func.EncodingNone, helper_func.EncodingHex, helper_func.EncodingBase64} {
//
//					t.Logf("%#v - %v", assignedProv, ty)
//
//					got, err := assignedProv.HashPasswordEncoding(ty, tt.args)
//					if (err != nil) != tt.wantErr {
//						t.Errorf("HashPasswordEncoding() error = %v, wantErr %v", err, tt.wantErr)
//						return
//					}
//					t.Log("1 HashPasswordEncoding():" + string(got))
//
//					time.Sleep(1)
//
//					got2, err := assignedProv.HashPasswordEncoding(ty, tt.args)
//					if (err != nil) != tt.wantErr {
//						t.Errorf("HashPasswordEncoding() error = %v, wantErr %v", err, tt.wantErr)
//						return
//					}
//					t.Log("2 HashPasswordEncoding():" + string(got2))
//
//					check0 := assignedProv.ComparePasswordEncoding(ty, tt.args+"1", got)
//					if check0 == nil {
//						t.Errorf("HashPasswordEncoding() error = %v, wantErr %v", err, tt.wantErr)
//					}
//
//					check01 := assignedProv.ComparePasswordEncoding(ty, tt.args[:len(tt.args)-1], got)
//					if check01 == nil {
//						t.Errorf("HashPasswordEncoding() error = %v, wantErr %v", err, tt.wantErr)
//					}
//
//					check := assignedProv.ComparePasswordEncoding(ty, tt.args, got)
//					if check != nil {
//						t.Errorf("HashPasswordEncoding() error = %v, wantErr %v", err, tt.wantErr)
//					}
//
//					check2 := assignedProv.ComparePasswordEncoding(ty, tt.args, got2)
//					if check2 != nil {
//						t.Errorf("HashPasswordEncoding() error = %v, wantErr %v", err, tt.wantErr)
//					}
//				}
//			}
//		})
//	}
//}

package handler

import (
	"gofiber/model"
	"reflect"
	"testing"
)

//func TestCreateBuildHashForm_Validate(t *testing.T) {
//
//	tests := []struct {
//		name        string
//		fields      CreateBuildHashForm
//		wantErrMsgs []string
//	}{
//		// TODO: Add test cases.
//		{"T1", CreateBuildHashForm{}, []string{
//			"Key: 'CreateBuildHashForm.BuildHashCore.Version' Error:Field validation for 'Version' failed on the 'required' tag",
//			"Key: 'CreateBuildHashForm.BuildHashCore.Hash' Error:Field validation for 'Hash' failed on the 'required' tag",
//		}},
//		{"T2", CreateBuildHashForm{Version: "1.2.3.4"}, []string{"Key: 'CreateBuildHashForm.BuildHashCore.Hash' Error:Field validation for 'Hash' failed on the 'required' tag"}},
//		{"T3", CreateBuildHashForm{Version: "1.2.3.4", Hash: "t123"}, []string{}},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if gotErrMsgs := tt.fields.Validate(); !reflect.DeepEqual(gotErrMsgs, tt.wantErrMsgs) {
//				t.Errorf("Validate() = %v, want %v", gotErrMsgs, tt.wantErrMsgs)
//			}
//		})
//	}
//}

func TestCreateBuildHashForm_ToCore(t *testing.T) {
	type fields struct {
		Version string
		Hash    string
		ObbHash []string
	}
	tests := []struct {
		name   string
		fields fields
		want   model.BuildHashCore
	}{
		// TODO: Add test cases.
		{"t0", fields{Version: "1.2.3", Hash: "hash"}, model.BuildHashCore{Version: "1.2.3", Hash: "hash", ObbHash: []byte("[]")}},

		{"t1", fields{Version: "1.2.3", Hash: "hash", ObbHash: []string{"ab", "cd"}}, model.BuildHashCore{Version: "1.2.3", Hash: "hash", ObbHash: []byte("[\"ab\",\"cd\"]")}},
		{"t1-2", fields{Version: "1.2.3", Hash: "hash", ObbHash: []string{"cd", "ab"}}, model.BuildHashCore{Version: "1.2.3", Hash: "hash", ObbHash: []byte("[\"ab\",\"cd\"]")}},
		{"t1-3", fields{Version: "1.2.3", Hash: "hash", ObbHash: []string{"cd", "efgh", "ab"}}, model.BuildHashCore{Version: "1.2.3", Hash: "hash", ObbHash: []byte("[\"ab\",\"cd\",\"efgh\"]")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CreateBuildHashForm{
				Version: tt.fields.Version,
				Hash:    tt.fields.Hash,
				ObbHash: tt.fields.ObbHash,
			}
			if got := c.ToCore(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToCore() = %v, want %v", got, tt.want)
			}
		})
	}
}

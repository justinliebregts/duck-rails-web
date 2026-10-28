package domain

import (
	"reflect"
	"testing"
)

func TestSHA256HashPROV_HashPasswordEncoding(t *testing.T) {

	tests := []struct {
		name    string
		args    string
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{"t0", "justin", "102cf10b5286bad9fcfe5e275ace3ddd7dcc23931fb0ca93dc223daf9877cabd", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SHA256HashPROV{}
			got, err := s.HashPasswordEncoding("hex", tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPasswordHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HashPasswordHex() got = %v, want %v", got, tt.want)
			}
		})
	}
}

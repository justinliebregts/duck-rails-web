package model

import (
	myDB "gofiber/database/mysql"
	"reflect"
	"testing"
	"time"
)

func TestGetPlatformUsageStatsFromDateRange(t *testing.T) {

	myDB.ConnectDb()

	now := CurrentTimestamp()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.AddDate(0, 0, 1).Add(-time.Second)

	type args struct {
		start time.Time
		end   time.Time
	}
	tests := []struct {
		name     string
		args     args
		wantPurs []PlatformCount
		wantErr  bool
	}{
		// TODO: Add test cases.
		{"t1", args{dayStart, dayEnd}, []PlatformCount{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPurs, err := GetPlatformUsageStatsFromDateRange(tt.args.start, tt.args.end, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPlatformUsageStatsFromDateRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPurs, tt.wantPurs) {
				t.Errorf("GetPlatformUsageStatsFromDateRange() gotPurs = %v, want %v", gotPurs, tt.wantPurs)
			}
		})
	}
}

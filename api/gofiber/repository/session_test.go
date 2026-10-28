package repository

import (
	"encoding/json"
	mySTO "gofiber/cache"
	myMDL "gofiber/model"
	"reflect"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2/utils"
)

// seedSession writes a session straight into Redis so the repository reads
// can be tested without the DB-backed session creation path.
func seedSession(t *testing.T) myMDL.Session {
	t.Helper()
	mySTO.ConnectDb()

	ses := myMDL.BuildSessionFromSessionBase(myMDL.FormSessionBase("z|repo_test_" + utils.UUIDv4()[:8]))

	sesBytes, err := ses.ToBytes(true)
	if err != nil {
		t.Fatal(err)
	}
	refBytes, err := json.Marshal(ses.Identity())
	if err != nil {
		t.Fatal(err)
	}

	keys := map[string][]byte{
		myMDL.CacheAccTknKey(ses.AccessToken):        sesBytes,
		myMDL.CacheRefTknKey(ses.RefreshToken):       refBytes,
		myMDL.CachePlayerUUIDKey(&ses.SessionBase):  []byte(ses.AccessToken),
	}
	for k, v := range keys {
		if err = mySTO.Create(k, v, time.Minute); err != nil {
			t.Skipf("local redis unavailable: %v", err)
		}
	}
	t.Cleanup(func() {
		for k := range keys {
			_ = mySTO.Delete(k)
		}
	})
	return ses
}

func Test_sessionRepositoryImpl_GetAccessTokenByUUID(t *testing.T) {
	ses := seedSession(t)

	tests := []struct {
		name    string
		sb      myMDL.SessionBase
		wantAt  string
		wantErr bool
	}{
		{"empty", myMDL.FormSessionBase(""), "", true},
		{"unknown", myMDL.FormSessionBase("z|repo_test_missing"), "", true},
		{"seeded", ses.SessionBase, ses.AccessToken, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sessionRepositoryImpl{}
			gotAt, err := s.GetAccessTokenByUUID(tt.sb)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccessTokenByUUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAt != tt.wantAt {
				t.Errorf("GetAccessTokenByUUID() gotAt = %v, want %v", gotAt, tt.wantAt)
			}
		})
	}
}

func Test_sessionRepositoryImpl_GetRefreshData(t *testing.T) {
	ses := seedSession(t)

	tests := []struct {
		name    string
		refTkn  string
		wantSb  myMDL.SessionBase
		wantErr bool
	}{
		{"empty", "", myMDL.SessionBase{}, true},
		{"fake", "a", myMDL.SessionBase{}, true},
		{"missing", utils.UUIDv4(), myMDL.SessionBase{}, true},
		{"seeded", ses.RefreshToken, ses.SessionBase, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sessionRepositoryImpl{}
			gotSb, err := s.GetRefreshData(tt.refTkn)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRefreshData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSb, tt.wantSb) {
				t.Errorf("GetRefreshData() gotSb = %v, want %v", gotSb, tt.wantSb)
			}
		})
	}
}

func Test_sessionRepositoryImpl_GetSession(t *testing.T) {
	ses := seedSession(t)

	tests := []struct {
		name    string
		accTkn  string
		wantSes myMDL.Session
		wantErr bool
	}{
		{"empty", "", myMDL.Session{}, true},
		{"fake", "a", myMDL.Session{}, true},
		{"miss", utils.UUIDv4(), myMDL.Session{}, true},
		{"seeded", ses.AccessToken, ses, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sessionRepositoryImpl{}
			gotSes, err := s.GetSession(tt.accTkn)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSes, tt.wantSes) {
				t.Errorf("GetSession() gotSes = %v, want %v", gotSes, tt.wantSes)
			}
		})
	}
}

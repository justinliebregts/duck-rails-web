package cache

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/storage/redis/v3"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestRedisNamespaceIsolation(t *testing.T) {
	server := miniredis.RunT(t)
	previousConfig, previousDB, previousMap := configDefault, DBConn, dBConnMap
	port, _ := strconv.Atoi(server.Port())
	configDefault = redis.Config{Host: server.Host(), Port: port}
	ConnectDb()
	t.Cleanup(func() {
		for _, db := range dBConnMap {
			_ = db.Close()
		}
		configDefault, DBConn, dBConnMap = previousConfig, previousDB, previousMap
	})
	for _, db := range []int{0, 1, 2} {
		t.Setenv("REDIS_NAMESPACE", "grc")
		if err := CreateWithDBIndex(db, "session/shared", []byte("classic"), time.Minute); err != nil {
			t.Fatal(err)
		}
		server.Select(db)
		if !server.Exists("grc/session/shared") {
			t.Fatal("Classic key compatibility lost")
		}
		t.Setenv("REDIS_NAMESPACE", "other-game")
		if _, err := FetchWithDBIndex(db, "session/shared"); err == nil {
			t.Fatal("read crossed namespace")
		}
		ok, err := CreateIfNotExistWithDBIndex(db, "session/shared", []byte("other"), time.Minute)
		if err != nil || !ok {
			t.Fatalf("namespace-local SETNX: %v %v", ok, err)
		}
		ok, err = CreateIfNotExistWithDBIndex(db, "session/shared", []byte("duplicate"), time.Minute)
		if err != nil || ok {
			t.Fatalf("duplicate SETNX: %v %v", ok, err)
		}
		keys, err := ScanWithDBIndex(db, "*")
		if err != nil || !reflect.DeepEqual(keys, []string{"session/shared"}) {
			t.Fatalf("scan leaked keys or physical prefix: %v %v", keys, err)
		}
		if err := DeleteWithDBIndex(db, "session/shared"); err != nil {
			t.Fatal(err)
		}
		t.Setenv("REDIS_NAMESPACE", "grc")
		value, err := FetchWithDBIndex(db, "session/shared")
		if err != nil || string(value) != "classic" {
			t.Fatalf("delete crossed namespace: %q %v", value, err)
		}
	}
	t.Setenv("REDIS_NAMESPACE", "other-game")
	if err := Create("session/shared", []byte("other"), time.Minute); err != nil {
		t.Fatal(err)
	}
	if value, err := Fetch("session/shared"); err != nil || string(value) != "other" {
		t.Fatalf("default fetch: %q %v", value, err)
	}
	if err := Delete("session/shared"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REDIS_NAMESPACE", "grc")
	if value, err := Fetch("session/shared"); err != nil || string(value) != "classic" {
		t.Fatal("default helpers crossed namespaces")
	}
}

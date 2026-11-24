package duckz

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"strings"
	"testing"
	"time"
)

func TestBoardRunID(t *testing.T) {
	id := strings.Repeat("a", 32)
	for _, raw := range []string{id, "run-" + id} {
		if got, ok := boardRunID(raw); !ok || got != id {
			t.Fatalf("unexpected canonical ID for %q: %q", raw, got)
		}
	}
	for _, raw := range []string{"run-", "run-run-" + id, "return-" + id} {
		if _, ok := boardRunID(raw); ok {
			t.Fatalf("accepted invalid ID %q", raw)
		}
	}
}

func TestBoardWeek(t *testing.T) {
	for date, want := range map[string]string{"2026-11-22T23:59:59Z": "2026-11-16", "2026-11-23T00:00:00Z": "2026-11-23", "2027-01-01T12:00:00Z": "2026-12-28"} {
		at, _ := time.Parse(time.RFC3339, date)
		if got := boardWeek(at); got != want {
			t.Fatalf("%s: %s", date, got)
		}
	}
}
func TestBoardDatabase(t *testing.T) {
	dsn := os.Getenv("RAIL_BOARD_TEST_DSN")
	if dsn == "" {
		t.Skip("requires disposable tails_test MySQL database")
	}
	if !strings.HasSuffix(dsn, "/tails_test") {
		t.Fatal("requires isolated tails_test schema")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile("../../../sql/rail_board.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, stmt := range strings.Split(string(migration), ";") {
		if strings.TrimSpace(stmt) != "" {
			if err = db.Exec(stmt).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, table := range []string{"rail_board_runs", "rail_board_totals"} {
			if err := tx.Exec("DELETE FROM " + table).Error; err != nil {
				return err
			}
		}
		week := "2026-11-16"
		r := boardReport{Run: strings.Repeat("a", 32), Name: "Waddles", Metrics: []boardMetric{{"attempt", 1}, {"meters", 400}, {"gold", 2}, {"coal", 3}, {"stop/quarry", 3}, {"down/bomber", 2}, {"hit/minion/bat", 7}, {"hit/minion/coal", 3}, {"escape", 1}}}
		for i := 0; i < 2; i++ {
			if err := applyBoard(tx, "duck-A", week, r); err != nil {
				return err
			}
		}
		r.Metrics = []boardMetric{{"meters", 250}, {"attempt", 1}, {"escape", 1}}
		if err := applyBoard(tx, "duck-A", week, r); err != nil {
			return err
		}
		r.Name = "Mallory"
		if err := applyBoard(tx, "duck-B", week, r); err != nil {
			return err
		}
		var s boardSnapshot
		if err := readBoard(tx, week, &s); err != nil {
			return err
		}
		if s.Attempts != 2 || s.Escapes != 2 || s.Meters != 650 || s.Gold != 2 || s.Coal != 3 || s.Stop.Amount != 3 || s.Deadliest.Label != "bomber" || s.Bonked.Amount != 10 || s.Item.Label != "bat" || len(s.Weekly) != 2 || s.Weekly[0].Label != "Waddles" {
			t.Fatalf("bad snapshot %+v", s)
		}
		var next boardSnapshot
		if err := readBoard(tx, "2026-11-23", &next); err != nil {
			return err
		}
		if len(next.Weekly) != 0 || next.Bonked.Amount != 0 || len(next.Lifetime) != 2 {
			t.Fatalf("invalid week rollover %+v", next)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

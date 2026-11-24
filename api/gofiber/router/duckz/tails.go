package duckz

import (
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	db "gofiber/database/mysql"
	mw "gofiber/middleware"
	"gorm.io/gorm"
)

type boardMetric struct {
	Key   string
	Value int
}
type boardReport struct {
	Run     string
	Name    string
	Metrics []boardMetric
}
type boardRow struct {
	Label  string `json:"label"`
	Amount int64  `json:"amount"`
}
type boardSnapshot struct {
	Attempts                      int64 `json:"attempts"`
	Escapes                       int64 `json:"escapes"`
	Coal                          int64 `json:"coal"`
	Gold                          int64 `json:"gold"`
	Meters                        int64 `json:"meters"`
	Stop, Deadliest, Bonked, Item boardRow
	Weekly, Lifetime              []boardRow
}

var runFormat = regexp.MustCompile("^[0-9a-f]{32}$")
var dimension = regexp.MustCompile("^[a-z0-9_-]{1,48}$")

func boardRunID(raw string) (string, bool) {
	id := strings.TrimPrefix(raw, "run-")
	return id, runFormat.MatchString(id)
}

func enemyKind(s string) bool {
	switch s {
	case "minion", "smasher", "bomber", "hook", "spider":
		return true
	}
	return false
}
func validMetric(key string, value int) bool {
	if value < 0 {
		return false
	}
	switch key {
	case "attempt", "escape":
		return value <= 1
	case "meters":
		return value <= 100000
	case "coal", "gold":
		return value <= 2000
	}
	parts := strings.Split(key, "/")
	if len(parts) == 2 && dimension.MatchString(parts[1]) {
		return value <= 2000 && (parts[0] == "stop" || parts[0] == "down" && enemyKind(parts[1]))
	}
	return len(parts) == 3 && parts[0] == "hit" && enemyKind(parts[1]) && dimension.MatchString(parts[2]) && value <= 50000
}
func boardWeek(t time.Time) string {
	t = t.UTC()
	return t.AddDate(0, 0, -(int(t.Weekday())+6)%7).Format("2006-01-02")
}
func addBoard(tx *gorm.DB, period, metric, subject, label string, delta int) error {
	return tx.Exec("INSERT INTO rail_board_totals (period,metric,subject,label,amount) VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE amount=amount+VALUES(amount),label=VALUES(label)", period, metric, subject, label, delta).Error
}
func applyBoard(tx *gorm.DB, player, week string, r boardReport) error {
	if err := tx.Exec("INSERT INTO rail_board_runs (player_id,run_id,counters) VALUES (?,?,JSON_OBJECT()) ON DUPLICATE KEY UPDATE run_id=VALUES(run_id)", player, r.Run).Error; err != nil {
		return err
	}
	var raw string
	if err := tx.Raw("SELECT counters FROM rail_board_runs WHERE player_id=? AND run_id=? FOR UPDATE", player, r.Run).Row().Scan(&raw); err != nil {
		return err
	}
	previous := map[string]int{}
	if err := json.Unmarshal([]byte(raw), &previous); err != nil {
		return err
	}
	sort.Slice(r.Metrics, func(i, j int) bool { return r.Metrics[i].Key < r.Metrics[j].Key })
	for _, m := range r.Metrics {
		delta := m.Value - previous[m.Key]
		if delta <= 0 {
			continue
		}
		previous[m.Key] = m.Value
		parts := strings.Split(m.Key, "/")
		if len(parts) == 1 {
			if err := addBoard(tx, "all", "total", m.Key, m.Key, delta); err != nil {
				return err
			}
			if m.Key == "escape" {
				for _, period := range []string{"all", week} {
					if err := addBoard(tx, period, "rank", player, r.Name, delta); err != nil {
						return err
					}
				}
			}
		} else {
			if err := addBoard(tx, week, parts[0], parts[1], parts[1], delta); err != nil {
				return err
			}
			if len(parts) == 3 {
				if err := addBoard(tx, week, "item/"+parts[1], parts[2], parts[2], delta); err != nil {
					return err
				}
			}
		}
	}
	if len(previous) > 256 {
		return errBoardSize
	}
	data, err := json.Marshal(previous)
	if err != nil {
		return err
	}
	return tx.Exec("UPDATE rail_board_runs SET counters=? WHERE player_id=? AND run_id=?", string(data), player, r.Run).Error
}

var errBoardSize = errors.New("too many counters")

func reportBoard(c *fiber.Ctx) error {
	var r boardReport
	if len(c.Body()) > 32768 || c.BodyParser(&r) != nil || len(r.Metrics) == 0 || len(r.Metrics) > 256 {
		return c.SendStatus(400)
	}
	var validRun bool
	if r.Run, validRun = boardRunID(r.Run); !validRun {
		return c.SendStatus(400)
	}
	if strings.ContainsAny(r.Name, "<>") || strings.IndexFunc(r.Name, unicode.IsControl) >= 0 || utf8.RuneCountInString(r.Name) > 80 {
		return c.SendStatus(400)
	}
	if strings.TrimSpace(r.Name) == "" {
		r.Name = "Duck"
	}
	for _, m := range r.Metrics {
		if !validMetric(m.Key, m.Value) {
			return c.SendStatus(400)
		}
	}
	player := mw.RetrieveSession(c)
	err := db.DBConn.Transaction(func(tx *gorm.DB) error { return applyBoard(tx, player.ComposedID(), boardWeek(time.Now()), r) })
	if errors.Is(err, errBoardSize) {
		return c.SendStatus(400)
	}
	if err != nil {
		return c.SendStatus(503)
	}
	return c.JSON(fiber.Map{"data": fiber.Map{}})
}
func readBoard(tx *gorm.DB, week string, s *boardSnapshot) error {
	rows := func(period, metric string, n int) ([]boardRow, error) {
		v := []boardRow{}
		err := tx.Raw("SELECT label,amount FROM rail_board_totals WHERE period=? AND metric=? ORDER BY amount DESC,subject ASC LIMIT ?", period, metric, n).Scan(&v).Error
		return v, err
	}
	totals, err := rows("all", "total", 5)
	if err != nil {
		return err
	}
	for _, v := range totals {
		switch v.Label {
		case "attempt":
			s.Attempts = v.Amount
		case "escape":
			s.Escapes = v.Amount
		case "coal":
			s.Coal = v.Amount
		case "gold":
			s.Gold = v.Amount
		case "meters":
			s.Meters = v.Amount
		}
	}
	if s.Weekly, err = rows(week, "rank", 10); err != nil {
		return err
	}
	if s.Lifetime, err = rows("all", "rank", 10); err != nil {
		return err
	}
	top := func(metric string, out *boardRow) error {
		v, e := rows(week, metric, 1)
		if len(v) > 0 {
			*out = v[0]
		}
		return e
	}
	if err = top("stop", &s.Stop); err != nil {
		return err
	}
	if err = top("down", &s.Deadliest); err != nil {
		return err
	}
	if err = top("hit", &s.Bonked); err != nil {
		return err
	}
	if s.Bonked.Amount > 0 {
		return top("item/"+s.Bonked.Label, &s.Item)
	}
	return nil
}

var boardCache struct {
	sync.Mutex
	week  string
	until time.Time
	value boardSnapshot
}

func getBoard(c *fiber.Ctx) error {
	boardCache.Lock()
	defer boardCache.Unlock()
	now := time.Now()
	week := boardWeek(now)
	if boardCache.week != week || !now.Before(boardCache.until) {
		var s boardSnapshot
		if err := db.DBConn.Transaction(func(tx *gorm.DB) error { return readBoard(tx, week, &s) }); err != nil {
			return c.SendStatus(503)
		}
		boardCache.value = s
		boardCache.week = week
		boardCache.until = now.Add(5 * time.Minute)
	}
	return c.JSON(fiber.Map{"data": boardCache.value})
}
func setupBoard(r fiber.Router) {
	r.Get("/rail-board", mw.CustomOAuth2Protected, getBoard)
	r.Post("/rail-board/report", mw.CustomOAuth2Protected, mw.NonceRepeatCheck(), reportBoard)
}

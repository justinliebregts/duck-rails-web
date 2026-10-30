package router

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"gofiber/database/mysql"
	"gofiber/middleware"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"regexp"
	"strings"
	"time"
)

// Diagnostic reports are separate from player moderation reports. Never expose the inbox to game clients.
type diagnosticRecord struct {
	ID         string    `json:"id" gorm:"primaryKey;size:32"`
	Reporter   string    `json:"-" gorm:"size:128;index"`
	CreatedAt  time.Time `json:"created"`
	Category   string    `json:"category" gorm:"size:100"`
	Summary    string    `json:"summary" gorm:"size:600"`
	Build      string    `json:"build" gorm:"size:100"`
	Platform   string    `json:"platform" gorm:"size:50"`
	Status     string    `json:"status" gorm:"size:16;index"`
	Resolution string    `json:"resolution" gorm:"size:1000"`
	Payload    string    `json:"payload,omitempty" gorm:"type:mediumtext"`
}

func (diagnosticRecord) TableName() string { return "diagnostic_reports" }

var diagnosticID = regexp.MustCompile(`^[a-f0-9]{32}$`)

type diagnosticInput struct {
	ID       string            `json:"id"`
	Created  string            `json:"created"`
	Category string            `json:"category"`
	Summary  string            `json:"summary"`
	Build    string            `json:"build"`
	Platform string            `json:"platform"`
	Snapshot json.RawMessage   `json:"snapshot"`
	History  []json.RawMessage `json:"history"`
	Events   []string          `json:"events"`
	Preset   json.RawMessage   `json:"preset"`
}

func validateDiagnostic(body []byte) (diagnosticInput, error) {
	var d diagnosticInput
	if len(body) > 1024*1024 {
		return d, errors.New("report exceeds 1 MiB")
	}
	if json.Unmarshal(body, &d) != nil || !diagnosticID.MatchString(d.ID) || len(d.Category) == 0 || len(d.Category) > 100 ||
		len(d.Summary) > 600 || len(d.Build) > 100 || len(d.Platform) > 50 || len(d.History) > 60 || len(d.Events) > 100 || len(d.Snapshot) == 0 {
		return d, errors.New("invalid diagnostic report")
	}
	for _, e := range d.Events {
		if len(e) > 1200 {
			return d, errors.New("event too long")
		}
	}
	if d.Created != "" {
		if _, err := time.Parse(time.RFC3339Nano, d.Created); err != nil {
			return d, errors.New("invalid capture timestamp")
		}
	}
	return d, nil
}

func SetupDiagnosticRoutes(r fiber.Router) {
	api := r.Group("/diagnostics", middleware.CustomOAuth2Protected)
	api.Post("/report", limiter.New(limiter.Config{Max: 20, Expiration: time.Minute}), submitDiagnostic)
	admin := api.Group("/admin", middleware.AdminPrivilegeCheckMiddleware)
	admin.Get("/reports", listDiagnostics)
	admin.Get("/reports/:id", getDiagnostic)
	admin.Post("/reports/:id/status", setDiagnosticStatus)
}
func diagnosticDB() (*gorm.DB, error) {
	if mysql.DBConn == nil {
		return nil, fiber.NewError(503, "Diagnostic database unavailable")
	}
	return mysql.DBConn, nil
}
func submitDiagnostic(c *fiber.Ctx) error {
	d, err := validateDiagnostic(c.Body())
	if err != nil {
		return fiber.NewError(400, err.Error())
	}
	db, err := diagnosticDB()
	if err != nil {
		return err
	}
	session := middleware.RetrieveSession(c)
	payload, _ := json.Marshal(d) // allowlisted top-level fields; ignore client status/upload flags
	record := diagnosticRecord{ID: d.ID, Reporter: session.ComposedID(), CreatedAt: time.Now().UTC(), Category: d.Category,
		Summary: d.Summary, Build: d.Build, Platform: d.Platform, Status: "open", Payload: string(payload)}
	// Idempotent retries must never reopen a resolved report or overwrite another reporter's evidence.
	result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&record)
	if result.Error != nil {
		return fiber.NewError(503, "Could not persist diagnostic report")
	}
	var existing diagnosticRecord
	if err := db.Select("id", "reporter").First(&existing, "id = ?", d.ID).Error; err != nil {
		return fiber.NewError(503, "Could not verify diagnostic receipt")
	}
	if existing.Reporter != record.Reporter {
		return fiber.NewError(409, "Report ID already used")
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"id": d.ID}})
}
func listDiagnostics(c *fiber.Ctx) error {
	status := c.Query("status", "open")
	if status != "open" && status != "resolved" && status != "all" {
		return fiber.NewError(400, "Invalid status")
	}
	db, err := diagnosticDB()
	if err != nil {
		return err
	}
	query := db.Model(&diagnosticRecord{})
	if status != "all" {
		query = query.Where("status = ?", status)
	}
	var count int64
	if query.Count(&count).Error != nil {
		return fiber.NewError(503, "Could not count reports")
	}
	var records []diagnosticRecord
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		return fiber.NewError(400, "Invalid offset")
	}
	if query.Omit("payload", "reporter").Order("created_at DESC, id DESC").Limit(100).Offset(offset).Find(&records).Error != nil {
		return fiber.NewError(503, "Could not read reports")
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"reports": records, "total": count}})
}
func getDiagnostic(c *fiber.Ctx) error {
	if !diagnosticID.MatchString(c.Params("id")) {
		return fiber.NewError(400, "Invalid ID")
	}
	db, err := diagnosticDB()
	if err != nil {
		return err
	}
	var record diagnosticRecord
	if err := db.First(&record, "id = ?", c.Params("id")).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fiber.ErrNotFound
	} else if err != nil {
		return fiber.NewError(503, "Could not read report")
	}
	return c.JSON(fiber.Map{"success": true, "data": record})
}
func setDiagnosticStatus(c *fiber.Ctx) error {
	var change struct {
		Status     string `json:"status"`
		Resolution string `json:"resolution"`
	}
	if !diagnosticID.MatchString(c.Params("id")) || len(c.Body()) > 4096 || c.BodyParser(&change) != nil ||
		(change.Status != "open" && change.Status != "resolved") || len(change.Resolution) > 1000 ||
		(change.Status == "resolved" && strings.TrimSpace(change.Resolution) == "") {
		return fiber.NewError(400, "Status and resolution evidence required")
	}
	db, err := diagnosticDB()
	if err != nil {
		return err
	}
	result := db.Model(&diagnosticRecord{}).Where("id = ?", c.Params("id")).Updates(map[string]interface{}{"status": change.Status, "resolution": change.Resolution})
	if result.Error != nil {
		return fiber.NewError(503, "Could not update report")
	}
	if result.RowsAffected == 0 {
		var count int64
		db.Model(&diagnosticRecord{}).Where("id = ?", c.Params("id")).Count(&count)
		if count == 0 {
			return fiber.ErrNotFound
		}
	}
	return c.JSON(fiber.Map{"success": true})
}

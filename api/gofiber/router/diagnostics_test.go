package router

import (
	"github.com/gofiber/fiber/v2"
	dbmysql "gofiber/database/mysql"
	"gofiber/model"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDiagnosticRetryDoesNotOverwriteEvidence(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: conn, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	before := dbmysql.DBConn
	dbmysql.DBConn = db
	defer func() { dbmysql.DBConn = before }()
	app := fiber.New()
	session := model.Session{SessionBase: model.NewSessionBase("OCULUS", "test-player")}
	app.Use(func(c *fiber.Ctx) error { c.Locals("player_session", session); return c.Next() })
	app.Post("/report", submitDiagnostic)
	id := "0123456789abcdef0123456789abcdef"
	for _, owner := range []string{session.ComposedID(), session.ComposedID(), "different-owner"} {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `diagnostic_reports` .* ON DUPLICATE KEY UPDATE `id`=`id`").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		mock.ExpectQuery("SELECT `id`,`reporter` FROM `diagnostic_reports`").WithArgs(id, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "reporter"}).AddRow(id, owner))
		req := httptest.NewRequest("POST", "/report", strings.NewReader(`{"id":"`+id+`","category":"manual","snapshot":{},"status":"resolved"}`))
		req.Header.Set("Content-Type", "application/json")
		response, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		expected := 200
		if owner != session.ComposedID() {
			expected = 409
		}
		if response.StatusCode != expected {
			t.Fatalf("retry status %d, expected %d", response.StatusCode, expected)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDiagnosticResolutionRequiresEvidence(t *testing.T) {
	app := fiber.New()
	app.Post("/:id/status", setDiagnosticStatus)
	req := httptest.NewRequest("POST", "/0123456789abcdef0123456789abcdef/status", strings.NewReader(`{"status":"resolved","resolution":" "}`))
	req.Header.Set("Content-Type", "application/json")
	response, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 400 {
		t.Fatal("resolved without evidence")
	}
}

func TestDiagnosticValidation(t *testing.T) {
	valid := `{"id":"0123456789abcdef0123456789abcdef","category":"manual","snapshot":{},"status":"resolved"}`
	if _, err := validateDiagnostic([]byte(valid)); err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{`{}`, strings.Replace(valid, "0123456789abcdef0123456789abcdef", "../escape", 1), strings.Repeat("x", 1024*1024+1), `{broken`} {
		if _, err := validateDiagnostic([]byte(body)); err == nil {
			t.Fatal("accepted invalid report")
		}
	}
}
func TestDiagnosticRoutesRejectAnonymous(t *testing.T) {
	app := fiber.New()
	SetupDiagnosticRoutes(app.Group("/api/v1"))
	for _, route := range []struct{ method, path string }{{"POST", "/report"}, {"GET", "/admin/reports"}, {"GET", "/admin/reports/0123456789abcdef0123456789abcdef"}, {"POST", "/admin/reports/0123456789abcdef0123456789abcdef/status"}} {
		response, err := app.Test(httptest.NewRequest(route.method, "/api/v1/diagnostics"+route.path, nil))
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 401 {
			t.Fatalf("%s returned %d", route.path, response.StatusCode)
		}
	}
}

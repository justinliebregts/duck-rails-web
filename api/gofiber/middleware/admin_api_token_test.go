package middleware

import (
	"github.com/gofiber/fiber/v2"
	"net/http/httptest"
	"testing"
)

func TestAdminAPIToken(t *testing.T) {
	t.Setenv("ADMIN_API_TOKEN", "test-admin-credential")
	for _, tc := range []struct {
		route, token string
		want         bool
	}{
		{"/api/v1/item/admin/all", "test-admin-credential", true},
		{"/api/v1/item/admin/all", "wrong", false},
		{"/api/v1/item/admin/all", "", false},
		{"/api/v1/session/:session_id", "test-admin-credential", false},
		{"/api/v1/player/:admin", "test-admin-credential", false},
	} {
		if got := acceptsAdminAPIToken(tc.route, tc.token); got != tc.want {
			t.Errorf("route %s: got %v want %v", tc.route, got, tc.want)
		}
	}
	t.Setenv("ADMIN_API_TOKEN", "")
	if acceptsAdminAPIToken("/api/v1/item/admin/all", "test-admin-credential") {
		t.Fatal("unconfigured token accepted")
	}
}

func TestAdminTokenInParentGroup(t *testing.T) {
	t.Setenv("ADMIN_API_TOKEN", "test-admin-credential")
	app := fiber.New()
	parent := app.Group("/api/v1/item", CustomOAuth2Protected)
	admin := parent.Group("/admin", AdminPrivilegeCheckMiddleware)
	admin.Get("/all", func(c *fiber.Ctx) error { return c.SendStatus(204) })
	req := httptest.NewRequest("GET", "/api/v1/item/admin/all", nil)
	req.Header.Set("Authorization", "test-admin-credential")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		t.Fatalf("got %d, want 204", resp.StatusCode)
	}
}
func TestAdminPathCannotComeFromParameter(t *testing.T) {
	app := fiber.New()
	app.Get("/api/v1/player/:feature/:id", func(c *fiber.Ctx) error {
		if isRegisteredAdminRequest(c) {
			t.Error("parameter treated as admin route")
		}
		return c.SendStatus(204)
	})
	req := httptest.NewRequest("GET", "/api/v1/player/admin/123", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

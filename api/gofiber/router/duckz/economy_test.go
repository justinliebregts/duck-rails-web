package duckz

import (
	"gofiber/config"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRailsEconomyRoutesAndAuthenticationScope(t *testing.T) {
	previous := config.DuckzEconomyJsonPath
	config.DuckzEconomyJsonPath = "../../config/json/duckz_economy.json"
	t.Cleanup(func() { config.DuckzEconomyJsonPath = previous })
	app := fiber.New()
	api := app.Group("/api/v1")
	SetupDuckzEconomyRoutes(api, nil)
	api.Get("/public-probe", func(c *fiber.Ctx) error { return c.SendStatus(204) })
	for _, route := range []struct{ method, path string }{
		{"GET", "/wallet"}, {"GET", "/store"}, {"POST", "/store/buy"},
		{"POST", "/coins/deposit"}, {"POST", "/iap/grant"}, {"POST", "/loadout"},
		{"GET", "/admin/wallet/test-player"}, {"POST", "/admin/coins/adjust"},
		{"GET", "/admin/transactions/test-player"}, {"GET", "/admin/check"},
	} {
		t.Run(route.method+route.path, func(t *testing.T) {
			response, err := app.Test(httptest.NewRequest(route.method, "/api/v1"+route.path, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != 401 {
				t.Fatalf("got %d, want authenticated endpoint (401)", response.StatusCode)
			}
		})
	}
	for path, want := range map[string]int{"/public-probe": 204, "/duckz/wallet": 404} {
		response, err := app.Test(httptest.NewRequest("GET", "/api/v1"+path, nil))
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != want {
			t.Fatalf("%s got %d, want %d", path, response.StatusCode, want)
		}
	}
}

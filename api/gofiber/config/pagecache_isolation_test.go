package config

import (
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http/httptest"
	"testing"
)

func TestPageCacheKeyIsolation(t *testing.T) {
	app := fiber.New()
	app.Get("/api/v1/info/all", func(c *fiber.Ctx) error { return c.SendString(pageCacheConfigDefault.KeyGenerator(c)) })
	for _, ns := range []string{"grc", "other-game"} {
		t.Setenv("REDIS_NAMESPACE", ns)
		res, err := app.Test(httptest.NewRequest("GET", "/api/v1/info/all", nil))
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != ns+"/pcache/api/v1/info/all" {
			t.Fatalf("unscoped page cache: %s", body)
		}
	}
}

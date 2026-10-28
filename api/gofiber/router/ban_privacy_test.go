package router

import (
	"net/http/httptest"
	"testing"

	myHDR "gofiber/handler"
	myREPO "gofiber/repository"
	mySRV "gofiber/service"
	myK8sSRV "gofiber/service/k8s"

	"github.com/gofiber/fiber/v2"
	"k8s.io/client-go/kubernetes/fake"
)

func TestBanRoutesDoNotExposePublicList(t *testing.T) {
	t.Chdir("..")
	app := fiber.New()
	k8s := fake.NewSimpleClientset()
	sessions := myREPO.NewSession()
	RegisterAPIRoutes(app.Group("/api/v1"), &AppContainer{
		CronJobSvc:  myK8sSRV.NewCronJobService(k8s),
		K8sClient:   k8s,
		SessionRepo: sessions,
	}, myHDR.NewSessionHandler(sessions, mySRV.NewSessionService(sessions)))

	own, admin := false, false
	for _, route := range app.GetRoutes(true) {
		if route.Path == "/api/v1/ban/all" || route.Path == "/api/v1/serviceban/all" {
			t.Fatalf("public ban-list route still registered: %s %s", route.Method, route.Path)
		}
		if route.Method == fiber.MethodGet && route.Path == "/api/v1/serviceban/get" {
			own = true
		}
		if route.Method == fiber.MethodGet && route.Path == "/api/v1/serviceban/admin/all" {
			admin = true
		}
	}
	if !own || !admin {
		t.Fatal("individual and admin ban routes must remain available")
	}

	for _, path := range []string{"/api/v1/serviceban/get", "/api/v1/serviceban/get?player_id=another-player", "/api/v1/serviceban/admin/all"} {
		res, err := app.Test(httptest.NewRequest(fiber.MethodGet, path, nil))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("%s: got %d, want 401", path, res.StatusCode)
		}
	}
}

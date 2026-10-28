package router

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	myHDR "gofiber/handler"
	myREPO "gofiber/repository"
	mySRV "gofiber/service"
	myK8sSRV "gofiber/service/k8s"

	"github.com/gofiber/fiber/v2"
	"k8s.io/client-go/kubernetes/fake"
)

// Live smoke test against a running local API (docker compose up -d in api/gofiber).
// Skipped when the API is not reachable. Override the target with
// SMOKE_HOST (default http://localhost:8090) and SMOKE_API_VERSION (default v99).
//
// It relies on the dev admin session the API creates when DEPLOY_ENV=development.

const smokeDevToken = "Xe94YkdkD9ejWG5rp42nugeHnEr"
const smokeDevPlayerID = "3724467550970486"

// Routes that call Meta, Cloudflare, Elasticsearch or K8s; locally these can
// legitimately fail, so their status is logged instead of asserted.
var smokeExternalRoutes = map[string]bool{
	"GET /player/v2/admin/oculus-org-scoped-id":                            true,
	"GET /player/v2/admin/oculus-attestation/list":                         true,
	"GET /player/v2/admin/oculus-attestation/ban-list":                     true,
	"GET /player/v2/admin/oculus-attestation/remote-ban-status/:unique_id": true,
	"GET /serviceban/admin/cloudflare-ip-blacklist":                        true,
	"GET /info/admin/namespaces/:ns/jobs/:jb/logs":                         true,
}

var smokeParamValues = map[string]string{
	"platform":        "oculus",
	"device_platform": "oculus",
	"app":             "gr",
	"feature":         "profile",
	"kind":            "badge",
	"ns":              "gunraiders-classic",
	"jb":              "smoke-missing-job",
	"cj":              "smoke-missing-cronjob",
	"player_name":     "smoke",
	"session_id":      "smoke",
}

var smokeParamRe = regexp.MustCompile(`:([A-Za-z_]+)(<[^>]+>)?`)

type smokeRoute struct {
	Method string
	Path   string // relative to /api/<version>, with :params
	URL    string // concrete path with params filled in
}

func smokeBaseURL(t *testing.T) (host, apiPrefix string) {
	host = strings.TrimRight(os.Getenv("SMOKE_HOST"), "/")
	if host == "" {
		host = "http://localhost:8090"
	}
	ver := os.Getenv("SMOKE_API_VERSION")
	if ver == "" {
		ver = "v99"
	}
	cl := http.Client{Timeout: 2 * time.Second}
	res, err := cl.Get(host + "/live")
	if err != nil {
		t.Skipf("local API not reachable at %s: %v", host, err)
	}
	_ = res.Body.Close()
	return host, host + "/api/" + ver
}

func smokeRoutes(t *testing.T) []smokeRoute {
	t.Helper()
	// Route setup loads config/json/* relative to the module root.
	t.Chdir("..")

	k8s := fake.NewSimpleClientset()
	sesRepo := myREPO.NewSession()
	app := fiber.New()
	RegisterAPIRoutes(app.Group("/api/smoke"), &AppContainer{
		CronJobSvc:  myK8sSRV.NewCronJobService(k8s),
		K8sClient:   k8s,
		SessionRepo: sesRepo,
	}, myHDR.NewSessionHandler(sesRepo, mySRV.NewSessionService(sesRepo)))

	var out []smokeRoute
	for _, r := range app.GetRoutes(true) {
		if r.Method == fiber.MethodHead {
			continue
		}
		rel := strings.TrimPrefix(r.Path, "/api/smoke")
		url := smokeParamRe.ReplaceAllStringFunc(rel, func(m string) string {
			name := smokeParamRe.FindStringSubmatch(m)[1]
			if v, ok := smokeParamValues[name]; ok {
				return v
			}
			if strings.HasSuffix(strings.ToLower(name), "id") || strings.Contains(name, "uuid") {
				if strings.Contains(strings.ToLower(name), "player") || name == "pla_id" || name == "platform_id" || name == "gr_uuid" || name == "unique_id" {
					return smokeDevPlayerID
				}
				return "999999999"
			}
			return "smoke"
		})
		out = append(out, smokeRoute{Method: r.Method, Path: rel, URL: url})
	}
	if len(out) == 0 {
		t.Fatal("no routes registered")
	}
	return out
}

func smokeDo(t *testing.T, method, url, token string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	cl := http.Client{Timeout: 15 * time.Second}
	res, err := cl.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 300))
	return res.StatusCode, string(body)
}

func isAdminPath(p string) bool {
	return strings.Contains(p, "/admin/") || strings.HasSuffix(p, "/admin")
}

// Every route exists on the running server, admin routes reject anonymous
// callers, and nothing 5xxs on an anonymous request with an empty body.
func TestLiveSmoke_AnonymousRequests(t *testing.T) {
	_, api := smokeBaseURL(t)
	for _, r := range smokeRoutes(t) {
		r := r
		t.Run(r.Method+" "+r.Path, func(t *testing.T) {
			status, body := smokeDo(t, r.Method, api+r.URL, "")
			switch {
			case status == 0:
				t.Fatalf("request failed: %s", body)
			case status >= 500:
				t.Errorf("status %d: %s", status, body)
			case isAdminPath(r.Path) && status != fiber.StatusUnauthorized && status != fiber.StatusForbidden:
				t.Errorf("admin route answered anonymous caller with %d: %s", status, body)
			case status == fiber.StatusNotFound && !strings.Contains(r.Path, ":"):
				t.Errorf("route missing on server (404): %s", body)
			}
		})
	}
}

// Every GET route answers the dev admin without a server error.
func TestLiveSmoke_AdminGets(t *testing.T) {
	_, api := smokeBaseURL(t)
	if status, body := smokeDo(t, fiber.MethodGet, api+"/info/admin/all", smokeDevToken); status != fiber.StatusOK {
		t.Skipf("dev admin session not usable (status %d: %s); is DEPLOY_ENV=development?", status, body)
	}
	for _, r := range smokeRoutes(t) {
		if r.Method != fiber.MethodGet {
			continue
		}
		r := r
		t.Run(r.Path, func(t *testing.T) {
			status, body := smokeDo(t, r.Method, api+r.URL, smokeDevToken)
			if smokeExternalRoutes[r.Method+" "+r.Path] {
				t.Logf("external dependency route: status %d", status)
				return
			}
			switch {
			case status == 0:
				t.Fatalf("request failed: %s", body)
			case status >= 500:
				t.Errorf("status %d: %s", status, body)
			case status == fiber.StatusUnauthorized || status == fiber.StatusForbidden:
				t.Errorf("dev admin rejected with %d: %s", status, body)
			case status == fiber.StatusNotFound && !strings.Contains(r.Path, ":"):
				t.Errorf("route missing on server (404): %s", body)
			}
		})
	}
}

func TestLiveSmoke_RouteCount(t *testing.T) {
	smokeBaseURL(t)
	n := len(smokeRoutes(t))
	t.Log(fmt.Sprintf("%d API routes registered", n))
}

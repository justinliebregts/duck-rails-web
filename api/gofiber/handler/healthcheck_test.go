package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthReadinessAndLiveness(t *testing.T) {
	old, oldHandler := EnablePrometheus, PromhttpHandlerFor
	t.Cleanup(func() { EnablePrometheus = old; PromhttpHandlerFor = oldHandler })
	for _, metrics := range []bool{false, true} {
		EnablePrometheus = metrics
		for _, failing := range []string{"", "database", "cache"} {
			check := func(name string) func() error {
				return func() error {
					if failing == name {
						return errors.New("unavailable")
					}
					return nil
				}
			}
			health := newHealthChecks(check("database"), check("cache"))
			ready := httptest.NewRecorder()
			health.ReadyEndpoint(ready, httptest.NewRequest(http.MethodGet, "/ready", nil))
			want := http.StatusOK
			if failing != "" {
				want = http.StatusServiceUnavailable
			}
			if ready.Code != want {
				t.Fatalf("metrics=%v failure=%s: readiness %d, want %d", metrics, failing, ready.Code, want)
			}
			live := httptest.NewRecorder()
			health.LiveEndpoint(live, httptest.NewRequest(http.MethodGet, "/live", nil))
			if live.Code != http.StatusOK {
				t.Fatalf("dependency failure must not fail liveness: %d", live.Code)
			}
		}
	}
}

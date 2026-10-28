package handler

import (
	"fmt"
	heaChk "github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gofiber/cache"
	myDB "gofiber/database/mysql"
	"net/http"
	"time"
)

var (
	EnablePrometheus   = true
	PromhttpHandlerFor http.Handler
)

// Liveness checks the serving process; readiness checks real dependencies.
// Dependency failures must not cause Kubernetes to restart a healthy process.
func newHealthChecks(database, redis func() error) heaChk.Handler {
	var health heaChk.Handler
	if EnablePrometheus {
		registry := prometheus.NewRegistry()
		health = heaChk.NewMetricsHandler(registry, "sample")
		PromhttpHandlerFor = promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	} else {
		health = heaChk.NewHandler()
	}
	health.AddReadinessCheck("database", database)
	health.AddReadinessCheck("cache", redis)
	return health
}

func HealthCheckHandler() (heaChk.Handler, error) {
	if myDB.DBConn == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}
	sqlDB, err := myDB.DBConn.DB()
	if err != nil {
		return nil, err
	}
	return newHealthChecks(heaChk.DatabasePingCheck(sqlDB, 3*time.Second), cache.HealthCheck), nil
}

package config

var (
//	PrometheusConfig = prometheus.Config{
//		DBName:          "db1",                       // use `DBName` as metrics label
//		RefreshInterval: 30,                          // Refresh metrics interval (default 15 seconds)
//		PushAddr:        "prometheus pusher address", // push metrics if `PushAddr` configured
//		StartServer:     true,                        // start http server to expose metrics
//		HTTPServerPort:  8080,                        // configure http server port, default port 8080 (if you have configured multiple instances, only the first `HTTPServerPort` will be used to start server)
//		MetricsCollector: []prometheus.MetricsCollector{
//			&prometheus.MySQL{
//				VariableNames: []string{"Threads_running"},
//			},
//		}, // user defined metrics
//	}
)

//func IsPrometheusDBEnabled() bool {
//	return os.Getenv("ENABLE_PROMETHEUS_DB") == "true"
//}

func SetupDBProm() {
	//if IsPrometheusDBEnabled() {
	//	myDB.DBConn.Use(prometheus.New(PrometheusConfig))
	//}
}

package mysql

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	cfgMDL "gofiber/model/config"

	"github.com/go-sql-driver/mysql"
	gormDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	//DSN refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	DSN string = "root:local_root_pass@tcp(localhost:3307)/gunraiders_classic?allowNativePasswords=false&checkConnLiveness=false&maxAllowedPacket=0&charset=utf8mb4&loc=Local&parseTime=True"
	/*
		NOTE:
		To handle time.Time correctly, you need to include parseTime as a parameter. (more parameters)
		To fully support UTF-8 encoding, you need to change charset=utf8 to charset=utf8mb4. See this article for a detailed explanation
	*/

	DBConn *gorm.DB

	GormDefaultConfig = gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	dsnSharedParams = map[string]string{"charset": "utf8mb4", "parseTime": "True"}
)

// BuildMySQLDSN build DSN string from myCFG.DBConfig
func BuildMySQLDSN(c cfgMDL.DBConfig) (dsn string) {
	config := mysql.Config{
		User:                 c.User,
		Passwd:               c.Password,
		Addr:                 c.Host,
		Net:                  c.Protocol,
		DBName:               c.DBName,
		AllowNativePasswords: true, //os.Getenv("DB_ALLOW_NATIVE_PASSWORDS") == "true",
		Loc:                  time.UTC,
		Params:               dsnSharedParams,
	}
	return config.FormatDSN()
}

func SetupGORMLogLevel(dev bool) {
	if dev {
		log.Printf("ConnectDbWithDSN Set GORM DebugMode")
		GormDefaultConfig.Logger = logger.Default.LogMode(logger.Info)
		return
	}

	log.Printf("ConnectDbWithDSN Set GORM WarnMode")
	GormDefaultConfig.Logger = logger.Default.LogMode(logger.Warn)
}

func NewDatabase(cfg cfgMDL.DBConfig) *gorm.DB {
	connectDbWithDSN(BuildMySQLDSN(cfg))
	DefineDbBoundary(DBConn)
	return DBConn
}

// connectDbWithDSN Connect with DB using assigned DSN string
func connectDbWithDSN(dsn string) {

	db, err := gorm.Open(gormDriver.Open(dsn), &GormDefaultConfig)

	if err != nil {
		log.Fatal("DSN: Failed to connect to database. \n", err)
		time.Sleep(time.Second)
		os.Exit(2)
	}

	log.Println("DSN: connected")
	//db.AutoMigrate(&models.Book{})
	DBConn = db
}

// envIntOr returns env var as int, or fallback when unset/invalid.
func envIntOr(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		log.Printf("mysql: invalid %s=%q, using %d", key, v, fallback)
		return fallback
	}
	return n
}

// DefineDbBoundary caps per-process MySQL pool size.
// Without this, database/sql defaults MaxOpenConns=0 (unlimited), which on TKE
// can exhaust TencentDB account max_user_connections (e.g. gruser @ 1024).
//
// Env overrides (optional):
//
//	DB_MAX_OPEN_CONNS      default 25 (0 falls back to 25; unlimited is not allowed)
//	DB_MAX_IDLE_CONNS      default 10
//	DB_CONN_MAX_LIFETIME_M minutes, default 30 (0 = no limit)
//	DB_CONN_MAX_IDLE_M     minutes, default 5  (0 = no limit)
func DefineDbBoundary(dbc *gorm.DB) {
	if dbc == nil {
		return
	}
	sqlDB, err := dbc.DB()
	if err != nil {
		log.Printf("DefineDbBoundary: sql.DB unavailable: %v", err)
		return
	}

	maxOpen := envIntOr("DB_MAX_OPEN_CONNS", 25)
	// database/sql treats MaxOpenConns=0 as unlimited; never allow that.
	if maxOpen == 0 {
		log.Printf("mysql: DB_MAX_OPEN_CONNS=0 disables the cap; falling back to 25")
		maxOpen = 25
	}
	maxIdle := envIntOr("DB_MAX_IDLE_CONNS", 10)
	lifeMin := envIntOr("DB_CONN_MAX_LIFETIME_M", 30)
	idleMin := envIntOr("DB_CONN_MAX_IDLE_M", 5)

	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	if lifeMin > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(lifeMin) * time.Minute)
	}
	if idleMin > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(idleMin) * time.Minute)
	}

	log.Printf("DefineDbBoundary: MaxOpen=%d MaxIdle=%d ConnMaxLifetime=%dm ConnMaxIdle=%dm",
		maxOpen, maxIdle, lifeMin, idleMin)
}

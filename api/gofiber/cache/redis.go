package cache

import (
	"context"
	"errors"
	myERR "gofiber/errors"
	"gofiber/internal/rediskeys"
	cfgMDL "gofiber/model/config"
	"sync"
	"time"

	"github.com/gofiber/storage/redis/v3"
)

var (
	mapLock sync.RWMutex

	DBConn    *redis.Storage
	dBConnMap map[int]*redis.Storage

	configDefault = redis.Config{
		Host:      "127.0.0.1",
		Port:      6380,
		Username:  "",
		Password:  "",
		URL:       "",
		Database:  0,
		Reset:     false,
		TLSConfig: nil,
	}
)

func DebugPrintConfig() redis.Config {
	return configDefault
}

func UpdateConfig(conf cfgMDL.CacheConfig) {
	configDefault.Host = conf.Host
	configDefault.Port = conf.Port
	configDefault.Username = conf.Username
	configDefault.Password = conf.Password
}

func DBConnSelector(dbIdx int) *redis.Storage {
	mapLock.RLock() // READ lock
	if val, ok := dBConnMap[dbIdx]; ok {
		mapLock.RUnlock()
		return val
	}
	mapLock.RUnlock() //READ unlock

	mapLock.Lock() // RW lock
	defer mapLock.Unlock()

	//
	if val, ok := dBConnMap[dbIdx]; ok {
		return val
	}

	dBConnMap[dbIdx] = connectWithIndex(dbIdx)
	return dBConnMap[dbIdx] // return DBConn
}

func connectWithIndex(dbIdx int) *redis.Storage {
	c0 := configDefault
	c0.Database = dbIdx
	return redis.New(c0)
}

func ConnectDb() {
	// Initialize custom config
	DBConn = redis.New(configDefault)

	dBConnMap = make(map[int]*redis.Storage)
	dBConnMap[0] = DBConn
	dBConnMap[1] = connectWithIndex(1)
}

func HealthCheck() error {
	if DBConn == nil {
		return myERR.ErrSTOConnectionNotExist
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := DBConn.Conn().Ping(ctx).Result()
	if err != nil {
		return errors.New("STO Response:" + err.Error())
	}

	return nil
}

func CreateIfNotExistWithDBIndex(db int, key string, content []byte, duration time.Duration) (bool, error) {
	dbc := DBConnSelector(db)
	return dbc.Conn().SetNX(context.Background(), rediskeys.Key(key), content, duration).Result()
}

func Create(key string, content []byte, duration time.Duration) error {
	return DBConn.Set(rediskeys.Key(key), content, duration)
}

func CreateWithDBIndex(db int, key string, content []byte, duration time.Duration) error {
	dbc := DBConnSelector(db)
	return dbc.Set(rediskeys.Key(key), content, duration)
}

func Fetch(key string) (byts []byte, err error) {
	if byts, err = DBConn.Get(rediskeys.Key(key)); err != nil {
		return
	}

	if len(byts) == 0 {
		err = myERR.ErrNotInCache
	}

	return
}

//func FetchHashWithDBIndex(db int, key, field string) (val string, err error) {
//	dbc := DBConnSelector(db)
//	if val, err = dbc.Conn().HGet(context.Background(), key, field).Result(); err != nil {
//		return
//	}
//
//	return
//}

func FetchWithDBIndex(db int, key string) (byts []byte, err error) {
	dbc := DBConnSelector(db)
	if byts, err = dbc.Get(rediskeys.Key(key)); err != nil {
		return
	}

	if len(byts) == 0 {
		err = myERR.ErrNotInCache
	}

	return
}

func Delete(key string) error {
	return DBConn.Delete(rediskeys.Key(key))
}

func DeleteWithDBIndex(db int, key string) error {
	dbc := DBConnSelector(db)
	return dbc.Delete(rediskeys.Key(key))
}

func ExistWithDBIndex(db int, key string) (exist bool, err error) {
	if _, err = FetchWithDBIndex(db, key); err != nil {
		if errors.Is(err, myERR.ErrNotInCache) {
			err = nil
		}
		return
	}

	exist = true
	return
}

//func Scan(pattern string) (keys []string) {
//	ctx := context.Background()
//	iter := DBConn.Conn().Scan(ctx, 0, pattern, 0).Iterator()
//
//	for iter.Next(ctx) {
//		keys = append(keys, iter.Val())
//
//		log.Printf("%v: %#v ", len(iter.Val()), iter.Val())
//	}
//
//	if err := iter.Err(); err != nil {
//		log.Printf("Scan Error:%#v", err)
//	}
//
//	return
//}

func Scan(pattern string) (keys []string, err error) {
	return ScanWithDBIndex(0, pattern)
}

func ScanWithDBIndex(db int, pattern string) (keys []string, err error) {
	dbc := DBConnSelector(db)

	ctx := context.Background()
	var cursor uint64
	for {
		var ke []string
		if ke, cursor, err = dbc.Conn().Scan(ctx, cursor, rediskeys.Key(pattern), 100).Result(); err != nil {
			break
		}

		for _, physical := range ke {
			keys = append(keys, rediskeys.Logical(physical))
		}
		if cursor == 0 {
			break
		}
		//log.Printf("scanned %d keys\n", len(ke))
	}
	return
}

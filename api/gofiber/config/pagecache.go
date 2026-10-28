package config

import (
	"fmt"
	"gofiber/internal/rediskeys"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/utils"
)

const (
	defaultExpirationSecond = 30
	pageCacheKeyPrefix      = "pcache"

	pathPatternPrefix = "^/?api/v[0-9]+/"

	pathPlatformPlatformIdPatternSuffix = "/?([a-z])+/([a-z][|])?([a-zA-Z][:][_])?[a-z0-9]{3,}$" // e.g. "oculus/123abc456def789gh0ij" or "google/g|123456" or "apple/a|t:123456"

	//OverwriteCacheKey = "PCacheKey_Overwrite"
)

var (
	// Paths (after /api/vNN/) whose GET/HEAD responses are cached in Redis.
	// Empty: response caching is off.
	specificCachedPages = []string{}

	defaultExpirationSecondString = fmt.Sprint(defaultExpirationSecond)

	pageCacheConfigDefault = cache.Config{
		Next:         nil,
		Expiration:   1 * time.Minute,
		CacheHeader:  "X-Cache",
		CacheControl: false,
		KeyGenerator: func(c *fiber.Ctx) string {
			//v, ok := c.GetReqHeaders()[OverwriteCacheKey]
			//if ok {
			//	log.Println("KeyGenerator use special key", v)
			//	return pageCacheKeyPrefix + v
			//}

			return rediskeys.Key(pageCacheKeyPrefix + utils.CopyString(c.Path()))
		},
		ExpirationGenerator:  nil,
		StoreResponseHeaders: false,
		Storage:              nil,
		MaxBytes:             0,
		Methods:              []string{fiber.MethodGet, fiber.MethodHead},
	}
)

func PrintCachedPageList() {
	log.Println("Caching Endpoints:")
	for _, val := range specificCachedPages {
		log.Println(val)
	}
}

// defaultNoStore `True` for skip Caching
func defaultNoStore(c *fiber.Ctx) bool {

	path := utils.CopyString(c.Path()) // "/api/v99/game/wholelist~"
	//single := strings.Split(path, "/")       // []string{"","api","v99","game","wholelist"}
	//section := strings.Join(single[3:], "/") // "game/wholelist~"

	// Call the new helper function
	if isPathCached(pathPatternPrefix, path, specificCachedPages) {
		return false // Don't skip caching
	}

	// Default: NoStore
	return true // Skipping Caching
}

func removePathPrefix(path string, prefixPattern string) string {
	pat := prefixPattern
	if !strings.HasPrefix(pat, "^") {
		pat = "^" + pat
	}

	rePrefix := regexp.MustCompile(pat)
	loc := rePrefix.FindStringIndex(path)

	if loc == nil {
		return ""
	}

	return path[loc[1]:]
}

// isPathCached checks if a path matches any of the patterns in the cache list.
func isPathCached(patternPrefix string, path string, cacheList []string) bool {
	// 1. Remove the static API version prefix from the path using the regex.
	pathWithoutPrefix := removePathPrefix(path, patternPrefix) // "api/v99/game/wholelist" to "game/wholelist"

	// This is the core logic from your original function.
	for _, val := range cacheList {

		var fullPattern string

		fullPattern = val
		if !strings.HasPrefix(fullPattern, "^") {
			fullPattern = "^" + fullPattern
		}
		if !strings.HasSuffix(fullPattern, "$") {
			fullPattern = fullPattern + "$"
		}

		//log.Printf("Matching path '%s' against pattern '%s'", pathWithoutPrefix, fullPattern)
		if matched, _ := regexp.MatchString(fullPattern, pathWithoutPrefix); matched {
			return true
		}
	}
	return false
}

func expirationGenerator(c *fiber.Ctx, cfg *cache.Config) time.Duration {

	querDurInt := c.QueryInt("Cache-Duration", -1)
	if querDurInt > 0 {
		// log.Println("Server Force Caching for", querDurInt, "s")
		return time.Second * time.Duration(querDurInt)
	}

	newCacheTime := defaultExpirationSecondString
	v, ok := c.GetReqHeaders()["Cache-Duration"]
	if ok {
		newCacheTime = strings.Join(v, ",")
	}

	if v2, ok2 := time.ParseDuration(newCacheTime + "s"); ok2 == nil {
		return v2
	}

	return defaultExpirationSecond
}

func ActivatePageCache(ap *fiber.App, cacheConnP fiber.Storage) {
	// Initialize default Page Cache config
	cnf := pageCacheConfigDefault
	cnf.Next = defaultNoStore
	cnf.CacheControl = true
	//cnf.StoreResponseHeaders = true // not really needed
	cnf.ExpirationGenerator = expirationGenerator

	//cnf.Storage = mySTO.CreateInMemoryStore()
	cnf.Storage = cacheConnP //mySTO.DBConn // assign redis as storage

	ap.Use(cache.New(cnf))
}

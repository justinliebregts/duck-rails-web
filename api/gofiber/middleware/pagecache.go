package middleware

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func PageCachedWithCacheDuration(cacheTimeInSecond int) func(c *fiber.Ctx) (err error) {
	return func(c *fiber.Ctx) (err error) {
		// TODO: TOO late to Add header at this moment
		// c.Request().Header.Add("Cache-Control", "cache")
		c.Request().Header.Add("Cache-Duration", strconv.Itoa(cacheTimeInSecond)) // 600 s is 10mins
		// log.Println("PageCachedWithCacheDuration")
		return c.Next()
	}
}

//func PageCachedWithSpecialCacheKey(cacheKey string) func(c *fiber.Ctx) (err error) {
//	return func(c *fiber.Ctx) (err error) {
//		c.Request().Header.Add(config.OverwriteCacheKey, cacheKey)
//		return c.Next()
//	}
//}

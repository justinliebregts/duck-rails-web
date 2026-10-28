package config

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
)

var eTagConfigDefault = etag.Config{
	Next: nil, //defaultETagCheck,
	Weak: false,
}

func EnableETag(app *fiber.App) {
	app.Use(etag.New(eTagConfigDefault))
}

//func defaultETagCheck(c *fiber.Ctx) bool {
//
//	//hea := c.GetReqHeaders()
//	//
//	//etagInResp := c.GetRespHeader(eTagHeader, "")
//	//log.Println("etagInResp:", etagInResp)
//	//
//	//if val, ok := hea[eTagHeader]; ok {
//	//	log.Println("etagInReq:", val)
//	//	if etagInResp == val {
//	//		return true
//	//	}
//	//}
//
//	return false
//}

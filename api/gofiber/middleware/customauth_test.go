package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	mySTO "gofiber/cache"
	"testing"
)

func TestCustomOAuth2Protected(t *testing.T) {

	mySTO.ConnectDb()

	app := fiber.New()

	c0 := app.AcquireCtx(&fasthttp.RequestCtx{})
	c0.Request().Header.Set("Autho", "Be")

	c1 := app.AcquireCtx(&fasthttp.RequestCtx{})
	c1.Request().Header.Set("Authorization", "test123")

	c2 := app.AcquireCtx(&fasthttp.RequestCtx{})
	c2.Request().Header.Set("Authorization", "Bearer t456")

	c3 := app.AcquireCtx(&fasthttp.RequestCtx{})
	c3.Request().Header.Set("Authorization", "Bearer t123456")

	tests := []struct {
		name    string
		args    *fiber.Ctx
		wantErr bool
	}{
		// TODO: Add test cases.
		//{"t0", c0, false},
		{"t1", c1, false},
		{"t2", c2, false},
		{"t3", c3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Logf("tt.args:%#v", tt.args.GetReqHeaders())

			if err := CustomOAuth2Protected(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("CustomOAuth2Protected() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				t.Logf("%#v", tt.args)
				//t.Logf("%#v", RetrieveSession(tt.args))
			}
		})
	}

	app.ReleaseCtx(c0)
	app.ReleaseCtx(c1)
	app.ReleaseCtx(c2)
	app.ReleaseCtx(c3)
}

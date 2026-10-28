package middleware

import (
	"github.com/gofiber/fiber/v2"
	oculus "gofiber/internal/service/remote/oculus"
	"gofiber/model"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type editorMetaStub struct {
	oculus.Client
	valid  bool
	org    string
	called bool
}

func (s *editorMetaStub) RequestOculusUserNonceValidate(q *model.UserNonceValidateQuery) (model.UserNonceValidateResponse, error) {
	s.called = true
	return model.UserNonceValidateResponse{IsValid: s.valid && q.UserID == "123" && q.Nonce == "proof"}, nil
}
func (s *editorMetaStub) GetOculusOrgScopedID(id string, q *model.GetOculusOrgScopedIDResponseQuery) (model.GetOculusOrgScopedIDResponse, error) {
	return model.GetOculusOrgScopedIDResponse{ID: id, ScopedID: s.org}, nil
}
func TestEditorLogin(t *testing.T) {
	for _, tc := range []struct {
		name, allow, org string
		valid            bool
		status           int
	}{
		{"disabled", "", "456", true, 401}, {"wrong account", "999", "456", true, 401},
		{"invalid proof", "456", "456", false, 401}, {"forged identity", "456", "999", true, 401},
		{"app-scoped ID is not accepted", "123", "456", true, 401},
		{"verified developer", " 456,789", "456", true, 204},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("META_EDITOR_USER_IDS", tc.allow)
			stub := &editorMetaStub{valid: tc.valid, org: tc.org}
			app := fiber.New()
			app.Post("/", MetaQuestAttestationMiddleware(stub), func(c *fiber.Ctx) error {
				if c.Locals(VerifiedRegistrationKey) == nil {
					t.Error("missing verified form")
				}
				return c.SendStatus(204)
			})
			form := url.Values{"platform_id": {"123"}, "player_id": {"456"}, "platform_token": {"proof"}, "platform": {"OCULUS"}}
			req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("X-Meta-Editor-Login", "1")
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tc.status {
				t.Fatalf("status %d want %d", resp.StatusCode, tc.status)
			}
			if (tc.allow == "") && stub.called {
				t.Error("disallowed user reached Meta")
			}
		})
	}
}

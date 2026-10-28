package middleware

import (
	"bytes"
	"encoding/json"
	"gofiber/model"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// Unity sends multipart fields, unlike the URL-encoded Editor middleware tests.
// Exercise the verified form and the response shape consumed by GameSessionClient.
func TestUnityMultipartEditorLogin(t *testing.T) {
	t.Setenv("META_EDITOR_USER_IDS", "456")
	stub := &editorMetaStub{valid: true, org: "456"}
	app := fiber.New()
	app.Post("/api/v1/player/v2/social-account/oculus", MetaQuestAttestationMiddleware(stub), func(c *fiber.Ctx) error {
		form, ok := c.Locals(VerifiedRegistrationKey).(*model.PlayerRegistrationForm)
		if !ok || form.PlayerID != "456" || form.PlatformID != "123" || form.Name != "Test Duck" || form.Region != "-" || form.DeviceName != "Unknown" {
			t.Fatalf("Unity registration fields did not survive middleware: %#v", form)
		}
		session := model.Session{
			SessionBase:    model.SessionBase{PlayerID: form.PlayerID, Platform: form.Platform},
			AccessToken:    "test-access-token",
			SessionRefresh: model.SessionRefresh{RefreshToken: "test-refresh-token"},
		}
		return c.JSON(fiber.Map{"data": session.FormResponseForRegister(form.Name)})
	})
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range map[string]string{
		"player_id": "456", "platform_id": "123", "platform_token": "proof",
		"platform": "OCULUS", "name": "Test Duck", "region": "-",
		"version": "0.2.7", "session": "duckrails", "device": "Unknown",
	} {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/api/v1/player/v2/social-account/oculus", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Meta-Editor-Login", "1")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var envelope struct {
		Data model.SessionResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	data := envelope.Data
	if data.AccessToken != "test-access-token" || data.RefreshToken != "test-refresh-token" || data.PlayerIdSeparated != "456" || data.Platform != "OCULUS" || data.DisplayName != "Test Duck" || data.GrId == "" || data.SessionId != -1 {
		t.Fatalf("incompatible Unity session response: %#v", data)
	}
}

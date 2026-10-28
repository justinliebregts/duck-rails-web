package middleware

import (
	"github.com/gofiber/fiber/v2"
	oculus "gofiber/internal/service/remote/oculus"
	"gofiber/model"
	"os"
	"strings"
)

// Editor login replaces device attestation only for explicitly allowed, Meta-verified users.
func verifyEditorLogin(c *fiber.Ctx, client oculus.Client) error {
	reject := func(message string) error {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": message})
	}
	var form model.PlayerRegistrationForm
	if err := c.BodyParser(&form); err != nil {
		return reject("Invalid Editor sign-in request.")
	}
	if strings.TrimSpace(os.Getenv("META_EDITOR_USER_IDS")) == "" {
		return reject("Editor login is disabled. Configure META_EDITOR_USER_IDS with organization-scoped Meta user IDs.")
	}
	if form.Platform != "OCULUS" || form.PlatformToken == "" {
		return reject("Meta user proof is required for Editor sign-in.")
	}
	verifier, ok := client.(interface {
		RequestOculusUserNonceValidate(*model.UserNonceValidateQuery) (model.UserNonceValidateResponse, error)
	})
	if !ok {
		return reject("Editor verification is unavailable.")
	}
	proof, err := verifier.RequestOculusUserNonceValidate(&model.UserNonceValidateQuery{UserID: form.PlatformID, Nonce: form.PlatformToken})
	if err != nil || !proof.IsValid {
		return reject("Meta rejected the Editor user proof. Check the Meta Platform test-user login.")
	}
	identity, err := client.GetOculusOrgScopedID(form.PlatformID, nil)
	if err != nil || identity.ID != form.PlatformID || identity.ScopedID == "" || identity.ScopedID != form.PlayerID {
		return reject("Meta Editor account identity could not be verified.")
	}
	allowed := false
	for _, id := range strings.Split(os.Getenv("META_EDITOR_USER_IDS"), ",") {
		if strings.TrimSpace(id) != "" && strings.TrimSpace(id) == identity.ScopedID {
			allowed = true
			break
		}
	}
	if !allowed {
		return reject("Editor account is not enabled. Add its organization-scoped Meta user ID to META_EDITOR_USER_IDS on the backend.")
	}
	c.Locals(VerifiedRegistrationKey, &form)
	return c.Next()
}

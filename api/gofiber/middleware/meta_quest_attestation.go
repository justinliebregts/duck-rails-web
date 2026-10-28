package middleware

import (
	authCFG "gofiber/config/remote_auth"
	myERR "gofiber/errors"
	ocuRmtSRV "gofiber/internal/service/remote/oculus"
	myMDL "gofiber/model"

	"github.com/gofiber/fiber/v2"
)

// MetaQuestAttestationMiddleware is only used in v2 /player/v2/social-account/oculus
func MetaQuestAttestationMiddleware(oculusRotSvc ocuRmtSRV.Client) fiber.Handler {
	return func(c *fiber.Ctx) (err error) {

		if c.Get("X-Meta-Editor-Login") == "1" {
			return verifyEditorLogin(c, oculusRotSvc)
		}

		if !authCFG.IsOculusAttestationEnabled() {
			return c.Next()
		}

		if authCFG.IsRemoteAuthSkipped(ExtractSkipAuthorization(c)) {
			return c.Next()
		}

		resp := myMDL.Default401Resp.Copy()

		usrInputPtr := new(myMDL.PlayerRegistrationForm)
		if err = c.BodyParser(usrInputPtr); err != nil {
			resp = myMDL.Default503Resp.Copy()
			resp.AddErrorAsMessages(err)
			return c.Status(resp.Status).JSON(resp)
		}

		var attestationToken string
		if attestationToken, err = ExtractBearerToken(c); err != nil {
			resp.SetErrorAndMessage(err.Error())
			return c.Status(resp.Status).JSON(resp)
		}

		//TODO: certain the nonce was in redis
		//var nonce string
		//if nonce, err = ExtractNonce(c); err != nil {
		//	resp.SetErrorAndMessage(err.Error())
		//	return c.Status(resp.Status).JSON(resp)
		//}
		//if len(nonce) == 0 {
		//	resp.SetErrorAndMessage("missing security nonce")
		//	return c.Status(resp.Status).JSON(resp)
		//}
		//c.IP() - nonce should be in Redis

		// OCULUS SERVICE: Cryptographic Verification
		q := ocuRmtSRV.NewVerifyAttestationTokenQuery(attestationToken)
		var oculusResp ocuRmtSRV.VerifyAttestationTokenResponse
		if oculusResp, err = oculusRotSvc.RequestOculusVerifyAttestationToken(q); err != nil {
			resp.SetErrorAndMessage(err.Error())
			return c.Status(resp.Status).JSON(resp)
		}

		_ = oculusResp.ParseClaims(false)

		if oculusResp.IsDeviceBanned() {
			//return myERR.ErrMetaQuestAttestationBanned
			resp.SetErrorAndMessage(myERR.ErrMetaQuestAttestationBanned.Error())
			return c.Status(resp.Status).JSON(resp)
		}

		// pass down usrInput
		c.Locals(VerifiedRegistrationKey, usrInputPtr)

		return c.Next()
	}
}

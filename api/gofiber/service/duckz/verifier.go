package duckz

import (
	"log"
	"os"

	myCFG "gofiber/config"
	myERR "gofiber/errors"
	oculus "gofiber/internal/service/remote/oculus"
	myMDL "gofiber/model"
)

const platformOculus = "OCULUS"

type metaEntitlementVerifier struct {
	client oculus.Client
}

// NewMetaEntitlementVerifier verifies durable SKUs with Meta S2S verify_entitlement.
// DUCKZ_IAP_SKIP_VERIFY=true bypasses the check, but only in development.
func NewMetaEntitlementVerifier(client oculus.Client) EntitlementVerifier {
	return &metaEntitlementVerifier{client: client}
}

func skipVerify() bool {
	return myCFG.IsEnvDevelopment() && os.Getenv("DUCKZ_IAP_SKIP_VERIFY") == "true"
}

// appScopedID maps the session's org-scoped id to this app's scoped id, which Meta S2S IAP calls need.
func (v *metaEntitlementVerifier) appScopedID(playerID, platform string) (string, error) {
	if platform != platformOculus || v.client == nil {
		return "", myERR.ErrDuckzUnsupportedPlatform
	}
	ids, err := v.client.GetOculusOrgScopedID(playerID, nil)
	if err != nil {
		log.Printf("[%s] app-scoped id lookup failed player=%s err=%v", logPrefix, playerID, err)
		return "", err
	}
	if ids.ID != "" {
		return ids.ID, nil
	}
	return playerID, nil
}

func (v *metaEntitlementVerifier) Verify(playerID, platform, sku string) error {
	if skipVerify() {
		log.Printf("[%s] iap verify skipped (dev) player=%s sku=%s", logPrefix, playerID, sku)
		return nil
	}
	appUserID, err := v.appScopedID(playerID, platform)
	if err != nil {
		return err
	}

	resp, err := v.client.RequestOculusVerifyItemOwnership(myMDL.VerifyItemOwnershipQuery{SKU: sku, UsrID: appUserID})
	if err != nil {
		return err
	}
	if !resp.Success {
		return myERR.ErrDuckzEntitlementNotOwned
	}
	return nil
}

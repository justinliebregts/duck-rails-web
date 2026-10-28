package oculus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	myCFG "gofiber/config/remote_auth"
	myERR "gofiber/errors"
	myMDL "gofiber/model"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	timeoutAppleRemoteApi = 30 * time.Second
)

type Client interface {
	DebugPrint() fiber.Map

	RequestOculusVerifyAttestationToken(VerifyAttestationTokenQuery) (VerifyAttestationTokenResponse, error)
	RequestOculusAttestationBanStatus(q *BanStatusRequest) (oculusResp BanStatusResponse, err error)
	RequestOculusAttestationBan(ctx context.Context, q *DeviceBanRequest) (oculusResp BanResponse, err error)

	//SaveToCache(*VerifyAttestationTokenResponse) error

	GetOculusOrgScopedID(oculusUsrID string, q *myMDL.GetOculusOrgScopedIDResponseQuery) (myMDL.GetOculusOrgScopedIDResponse, error)
	RequestOculusVerifyItemOwnership(vq myMDL.VerifyItemOwnershipQuery) (myMDL.OCULUSResponseBase, error)
	RequestOculusConsumeIAPItem(q myMDL.OculusConsumeIAPItemQuery) (myMDL.OCULUSResponseBase, error)
}

func NewOculusService(conf *myCFG.OculusConfig) Client {
	return &oculusClient{
		config:     conf,
		httpClient: &http.Client{Timeout: timeoutAppleRemoteApi},
		//certCache:  sync.Map{},
	}
}

type oculusClient struct {
	config     *myCFG.OculusConfig
	httpClient *http.Client
}

func (c *oculusClient) buildVerifyItemOwnershipQuery(vq myMDL.VerifyItemOwnershipQuery) url.Values {
	params := url.Values{}
	params.Add("sku", vq.SKU)
	params.Add("access_token", c.config.PlatformAccessToken)
	if vq.UsrID != "" {
		params.Add("user_id", vq.UsrID)
	}
	return params
}

// RequestOculusVerifyItemOwnership calls Meta S2S verify_entitlement; UsrID must be the app-scoped user id.
func (c *oculusClient) RequestOculusVerifyItemOwnership(vq myMDL.VerifyItemOwnershipQuery) (OculusResp myMDL.OCULUSResponseBase, err error) {
	var req *http.Request

	body := c.buildVerifyItemOwnershipQuery(vq).Encode()

	if req, err = http.NewRequest("POST", fmt.Sprintf("%v/%v%v?%v", strings.TrimRight(c.config.PlatformServer, "/"), c.config.OcCred.AppID, VerifyItemOwnershipUrl, body), nil); err != nil {
		return
	}
	log.Printf("RequestOculusVerifyItemOwnership : sku=%s user=%s", vq.SKU, vq.UsrID)

	// Send request
	var resp *http.Response
	if resp, err = c.httpClient.Do(req); err != nil {
		return
	}

	defer resp.Body.Close()

	var respBytes []byte
	if respBytes, err = io.ReadAll(resp.Body); err != nil {
		return
	}

	if err = json.Unmarshal(respBytes, &OculusResp); err != nil {
		return
	}

	if len(OculusResp.Error.Message) > 0 {
		err = errors.New(OculusResp.Error.Message)
	}

	return
}

func (c *oculusClient) RequestOculusConsumeIAPItem(q myMDL.OculusConsumeIAPItemQuery) (OculusResp myMDL.OCULUSResponseBase, err error) {
	var req *http.Request

	body := c.buildVerifyItemOwnershipQuery(q.VerifyItemOwnershipQuery).Encode()

	if req, err = http.NewRequest("POST", fmt.Sprintf("%v/%v%v?%v", strings.TrimRight(c.config.PlatformServer, "/"), c.config.OcCred.AppID, ConsumeIAPItemUrl, body), nil); err != nil {
		return
	}
	log.Printf("RequestOculusConsumeIAPItem : sku=%s user=%s", q.SKU, q.UsrID)

	// Send request
	var resp *http.Response
	if resp, err = c.httpClient.Do(req); err != nil {
		return
	}

	defer resp.Body.Close()

	var respBytes []byte
	if respBytes, err = io.ReadAll(resp.Body); err != nil {
		return
	}

	if err = json.Unmarshal(respBytes, &OculusResp); err != nil {
		return
	}

	if len(OculusResp.Error.Message) > 0 {
		err = errors.New(OculusResp.Error.Message)
	}

	return
}

func (c *oculusClient) buildRetrieveItemsOwnedQuery(rq myMDL.RetrieveItemsOwnedQuery) url.Values {
	params := url.Values{}
	params.Add("access_token", c.config.PlatformAccessToken) //params.Add("access_token", r.AccessToken)
	params.Add("user_id", rq.OrgScopedID)
	params.Add("fields", strings.Join(rq.Fields, ","))

	//log.Println(params)
	return params
}

func (c *oculusClient) RequestOculusRetrieveItemsOwned(q myMDL.RetrieveItemsOwnedQuery) (oculusResp myMDL.RetrieveItemsOwnedResponse, err error) {
	var req *http.Request

	body := c.buildRetrieveItemsOwnedQuery(q).Encode()

	//https://developer.oculus.com/documentation/unity/ps-iap-s2s/
	// OculusPlatformServer+RetrieveItemsOwnedUrl+"?"+q.BuildQuery().Encode()
	if req, err = http.NewRequest("GET", fmt.Sprintf("%v/%v%v?%v", c.config.PlatformServer, c.config.OcCred.AppID, RetrieveItemsOwnedUrl, body), nil); err != nil {
		return
	}
	//log.Printf("RequestOculusRetrieveItemsOwned : %v", req.URL)

	// Send request
	var resp *http.Response
	if resp, err = c.httpClient.Do(req); err != nil {
		return
	}

	defer resp.Body.Close()

	var respBytes []byte
	if respBytes, err = io.ReadAll(resp.Body); err != nil {
		return
	}

	if err = json.Unmarshal(respBytes, &oculusResp); err != nil {
		return
	}

	if len(oculusResp.Error.Message) > 0 {
		err = errors.New(oculusResp.Error.Message)
	}

	return
}

func (c *oculusClient) RequestOculusUserNonceValidate(q *myMDL.UserNonceValidateQuery) (OculusResp myMDL.UserNonceValidateResponse, err error) {
	var req *http.Request

	body := c.BuildParameter(q)

	// Combine the base URL, path, and parameters
	fullURL := fmt.Sprintf("%s%s?%s", c.config.PlatformServer, UserNonceValidateUrl, body)

	if req, err = http.NewRequest("POST", fullURL, nil); err != nil {
		return
	}
	//log.Printf("RequestOculusUserNonceValidate : %v", req.URL)

	// Send request
	//client := http.Client{}
	//if q.RequestTimeout != 0 {
	//	client.Timeout = q.RequestTimeout // Timeout set to 5 seconds by default
	//}

	var resp *http.Response
	if resp, err = c.httpClient.Do(req); err != nil {
		return
	}

	defer resp.Body.Close()

	var respBytes []byte
	if respBytes, err = io.ReadAll(resp.Body); err != nil {
		return
	}

	//log.Printf("%#v", string(respBytes))
	if err = json.Unmarshal(respBytes, &OculusResp); err != nil {
		//log.Println(string(respBytes))
		return
	}

	if !OculusResp.IsValid {
		err = myERR.NoFound
	}

	return
}

func (c *oculusClient) BuildQuery(r *myMDL.GetOculusOrgScopedIDResponseQuery) url.Values {
	params := url.Values{}
	params.Add("access_token", c.config.PlatformAccessToken) //r.AccessToken)
	params.Add("fields", strings.Join(r.Fields, ","))

	log.Println(params)
	return params
}

// GetOculusOrgScopedID from Oculus usrID to Oculus Verified Org Scoped ID
// the ID is actually desired value; input oculusUsrID will be in GetOculusOrgScopedIDResponse.ScopedID
func (c *oculusClient) GetOculusOrgScopedID(oculusUsrID string, q *myMDL.GetOculusOrgScopedIDResponseQuery) (respOrgScopedID myMDL.GetOculusOrgScopedIDResponse, err error) {
	//https://developer.oculus.com/documentation/unity/ps-ownership#retrieve-a-verified-org-scoped-id

	if q == nil {
		q = &myMDL.GetOculusOrgScopedIDResponseQuery{Fields: []string{"id", "alias", "org_scoped_id"}}
	}

	reqURL, err := url.Parse(fmt.Sprintf("%s/%s", strings.TrimRight(c.config.PlatformServer, "/"), oculusUsrID))
	if err != nil {
		return respOrgScopedID, fmt.Errorf("invalid oculus platform server url: %w", err)
	}
	reqURL.RawQuery = c.BuildQuery(q).Encode()

	var req *http.Request
	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return respOrgScopedID, fmt.Errorf("failed to create http request: %w", err)
	}
	//log.Printf("GetOculusOrgScopedID : %v", req.URL)

	// Send request
	var resp *http.Response
	if resp, err = c.httpClient.Do(req); err != nil {
		err = fmt.Errorf("oculus api request failed: %w", err)
		return
	}
	defer resp.Body.Close()

	var respBytes []byte
	if respBytes, err = io.ReadAll(resp.Body); err != nil {
		err = fmt.Errorf("failed to read oculus response body: %w", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("oculus api error (status %d): %s", resp.StatusCode, string(respBytes))
		return
	}

	//log.Printf("GetOculusOrgScopedID : %v", string(respBytes))

	if err = json.Unmarshal(respBytes, &respOrgScopedID); err != nil {
		err = fmt.Errorf("failed to unmarshal oculus response: %w", err)
		return
	}

	if !respOrgScopedID.IsValid() {
		err = myERR.NoPlayerMatch
	}
	return
}

func (c *oculusClient) BuildUserNonceValidateQuery(u *myMDL.UserNonceValidateQuery, uid string, nonce string, withTimeout bool) {
	u.UserID = uid
	u.Nonce = nonce

	if !withTimeout {
		u.RequestTimeout = requestTimeout
	}
}

func (c *oculusClient) BuildParameter(u *myMDL.UserNonceValidateQuery) string {
	parameters := url.Values{}

	// Add parameters to the URL
	parameters.Add("access_token", c.config.PlatformAccessToken)
	parameters.Add("user_id", u.UserID)
	parameters.Add("nonce", u.Nonce)

	return parameters.Encode()
}

// DebugPrint returns debug information about Oculus configuration
func (c *oculusClient) DebugPrint() fiber.Map {
	return fiber.Map{
		"oculus_config":             c.config.DebugPrint(),
		"ENABLE_OCULUS_ATTESTATION": myCFG.IsOculusAttestationEnabled(),
	}
}

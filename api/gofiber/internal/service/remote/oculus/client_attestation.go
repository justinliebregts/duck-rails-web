package oculus

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	myCFG "gofiber/config"
	myERR "gofiber/errors"
	myMDL "gofiber/model"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	VerifyAttestationTokenPath = "/platform_integrity/verify"
	BanWithAttestationPath     = "/platform_integrity/device_ban"
	BanStatusPath              = "/platform_integrity/device_ban_status"
)

type VerifyAttestationTokenQuery struct {
	AttestationToken string `json:"attestation_token"`
}

func (tq *VerifyAttestationTokenQuery) formUrl(platformServer string, accessToken string) string {
	//https://graph.oculus.com/platform_integrity/verify?token=<attestation_token>&access_token=<access_token>

	// Join the base URL and the path safely using url.JoinPath.
	urlOnly, err := url.JoinPath(platformServer, VerifyAttestationTokenPath)
	if err != nil {
		panic(err) // Handle error appropriately
	}

	// Create a new URL object.
	u, err := url.Parse(urlOnly)
	if err != nil {
		panic(err) // Handle error appropriately
	}

	// Create a url.Values map to hold query parameters. This automatically handles URL encoding.
	q := u.Query()
	q.Add("token", tq.AttestationToken)
	q.Add("access_token", accessToken)

	// Encode the query parameters and assign them to the URL.
	u.RawQuery = q.Encode()

	myCFG.DebugLogPrintf("formUrl: %+v", u.String())
	return u.String()
}

func NewVerifyAttestationTokenQuery(attestationToken string) VerifyAttestationTokenQuery {
	return VerifyAttestationTokenQuery{AttestationToken: attestationToken}
}

type VerifyAttestationTokenResponse struct {
	Data []struct {
		Message *string `json:"message" enums:"success,invalid signature,token expired"`
		// Claims is a Base64-encoded JWS string containing DeviceId, Nonce, and Integrity status.
		ClaimsInBase64 *string `json:"claims" extensions:"x-format=base64"`
	} `json:"data"`

	Claim *myMDL.AttestationClaims `json:"-"`
}

func (v *VerifyAttestationTokenResponse) IsDeviceBanned() bool {

	if len(v.Data) == 0 || v.Data[0].Message == nil {
		// No data means we can't verify, treat as banned (or handle differently)
		return true
	}

	// Assuming we only care about the first element in `data`
	if !strings.EqualFold(*v.Data[0].Message, "success") {
		return true
	}

	// now check with claim
	if parseErr := v.ParseClaims(false); parseErr == nil {
		if v.Claim == nil {
			return true
		}
		return v.Claim.IsDeviceBanned()
	}

	return true
}

func (v *VerifyAttestationTokenResponse) ParseClaims(force bool) error {
	if v.Claim != nil && !force {
		myCFG.DebugLogPrintf("ParseClaims ALREADY : %+v", v.Claim)
		return nil
	}

	if len(v.Data) == 0 {
		return myERR.NoFound // define this error somewhere
	}

	raw := *v.Data[0].ClaimsInBase64
	// Base64URL decode
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return err
	}

	var claims myMDL.AttestationClaims
	if err = json.Unmarshal(decoded, &claims); err != nil {
		return err
	}

	v.Claim = &claims
	myCFG.DebugLogPrintf("ParseClaims : %+v", claims)
	return nil
}

func (c *oculusClient) RequestOculusVerifyAttestationToken(q VerifyAttestationTokenQuery) (oculusResp VerifyAttestationTokenResponse, err error) {
	// https://developers.meta.com/horizon/documentation/unity/ps-attestation-api

	var req *http.Request

	if req, err = http.NewRequest("GET", q.formUrl(c.config.PlatformServer, c.config.PlatformAccessToken), nil); err != nil {
		return
	}

	myCFG.DebugLogPrintf("RequestOculusVerifyAttestationToken : %v", req.URL)

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
	myCFG.DebugLogPrintf("RequestOculusVerifyAttestationToken resp : %q", respBytes)
	err = json.Unmarshal(respBytes, &oculusResp)

	return
}

type BanStatusRequest struct {
	UniqueId string `json:"unique_id"`
	BanId    string `json:"ban_id"`
}

func (bsr *BanStatusRequest) HasBanId() bool {
	// Returns true if BanId is not an empty string
	return bsr.BanId != ""
}

func (bsr *BanStatusRequest) formUrl(platformServer string, accessToken string) string {
	//https://graph.oculus.com/platform_integrity/device_ban_status?
	//unique_id =<unique_id>&
	//access_token=<access_token>

	// Join the base URL and the path safely using url.JoinPath.
	urlOnly, err := url.JoinPath(platformServer, BanStatusPath)
	if err != nil {
		panic(err) // Handle error appropriately
	}

	// Create a new URL object.
	u, err := url.Parse(urlOnly)
	if err != nil {
		panic(err) // Handle error appropriately
	}

	// Create a url.Values map to hold query parameters. This automatically handles URL encoding.
	q := u.Query()
	q.Add("access_token", accessToken)

	if bsr.HasBanId() {
		q.Add("ban_id", bsr.BanId)
	} else {
		q.Add("unique_id", bsr.UniqueId)
	}

	// Encode the query parameters and assign them to the URL.
	u.RawQuery = q.Encode()

	myCFG.DebugLogPrintf("formUrl: %+v", u.String())
	return u.String()
}

type BanStatusResponse struct {
	Message      string          `json:"message,omitempty"`
	Data         []BanStatusData `json:"data,omitempty"`
	Error        *OCApiError     `json:"error,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
}

func (r *BanStatusResponse) Result() string {
	// 1. Handle API Errors
	if r.Error != nil {
		switch r.Error.ErrorSubcode {
		case 1614026:
			return "invalid unique_id: record not found or expired (IDs change every 30 days)"
		case 1614027:
			return "unauthorized: unique_id does not match the App ID in your access token"
		default:
			return r.Error.Message
		}
	}

	// 2. Handle Success (Check if data slice is not empty)
	if len(r.Data) > 0 {
		status := r.Data[0]
		if status.IsBanned {
			return fmt.Sprintf("Player is BANNED. Remaining: %s minutes", status.RemainingBanTime)
		}
		return "Player is not banned."
	}

	return "No status data available"
}

type BanStatusData struct {
	Message          string `json:"message,omitempty"`
	IsBanned         bool   `json:"is_banned"`
	RemainingBanTime string `json:"remaining_ban_time,omitempty"`
	BanCreationDate  string `json:"ban_creation_date,omitempty"`
}

func (c *oculusClient) RequestOculusAttestationBanStatus(q *BanStatusRequest) (oculusResp BanStatusResponse, err error) {
	// https://developers.meta.com/horizon/documentation/spatial-sdk/ps-attestation-api/#how-to-ban-a-device

	var req *http.Request

	if req, err = http.NewRequest("GET", q.formUrl(c.config.PlatformServer, c.config.PlatformAccessToken), nil); err != nil {
		return
	}

	myCFG.DebugLogPrintf("RequestOculusAttestationBanStatus : %v", req.URL)

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
	myCFG.DebugLogPrintf("RequestOculusAttestationBanStatus resp : %q", respBytes)
	err = json.Unmarshal(respBytes, &oculusResp)

	return
}

type DeviceBanRequest struct {
	myMDL.DeviceBan
	UniqueId string `json:"unique_id"`
	BanId    string `json:"ban_id"`
}

func (dbr *DeviceBanRequest) HasBanId() bool {
	// Returns true if BanId is not an empty string
	return dbr.BanId != ""
}

func (dbr *DeviceBanRequest) formUrl(platformServer string, accessToken string) string {
	//https://graph.oculus.com/platform_integrity/device_ban?
	//method=POST&
	//unique_id=<unique_id>&
	//is_banned=<is_banned>&
	//remaining_time_in_minute=<remaining_time_in_minute>&
	//access_token=<access_token>

	// Join the base URL and the path safely using url.JoinPath.
	urlOnly, err := url.JoinPath(platformServer, BanWithAttestationPath)
	if err != nil {
		panic(err) // Handle error appropriately
	}

	// Create a new URL object.
	u, err := url.Parse(urlOnly)
	if err != nil {
		panic(err) // Handle error appropriately
	}

	// Create a url.Values map to hold query parameters. This automatically handles URL encoding.
	q := u.Query()
	q.Add("method", "POST")
	q.Add("is_banned", dbr.IsCurrentlyBanned())
	q.Add("access_token", accessToken)

	if dbr.HasBanId() {
		q.Add("ban_id", dbr.BanId)
	} else {
		q.Add("unique_id", dbr.UniqueId)
	}

	if dbr.IsBanned {
		q.Add("remaining_time_in_minute", fmt.Sprintf("%d", dbr.RemainingBanTime))
	}

	// Encode the query parameters and assign them to the URL.
	u.RawQuery = q.Encode()

	myCFG.DebugLogPrintf("formUrl: %+v", u.String())
	return u.String()
}

type OCApiError struct {
	Message      string                 `json:"message"`
	Type         string                 `json:"type"`
	Code         int                    `json:"code"`
	ErrorSubcode int                    `json:"error_subcode"`
	FBTraceID    string                 `json:"fbtrace_id"`
	ErrorData    map[string]interface{} `json:"error_data"`
}

type BanResponse struct {
	Message string      `json:"message,omitempty"`
	BanID   string      `json:"ban_id,omitempty"` // Empty string on reversal, populated on update
	Error   *OCApiError `json:"error,omitempty"`
}

func (r *BanResponse) Result() string {
	// 1. Priority: Handle Errors
	if r.Error != nil {
		switch r.Error.ErrorSubcode {
		case 1614033:
			return "ban record not found or already reversed"
		case 1614036:
			return "unauthorized: access token mismatch"
		default:
			return r.Error.Message
		}
	}

	// 2. Priority: Specific Success Scenarios
	if r.BanID != "" {
		return fmt.Sprintf("Ban active. New ID: %s", r.BanID)
	}

	// Since the Oculus docs say ban_id is "" on reversal:
	if r.Message == "Success" {
		return "Ban successfully reversed."
	}

	// 3. Fallback
	return "Unknown response state"
}

func (c *oculusClient) RequestOculusAttestationBan(ctx context.Context, q *DeviceBanRequest) (oculusResp BanResponse, err error) {
	// https://developers.meta.com/horizon/documentation/spatial-sdk/ps-attestation-api/#how-to-ban-a-device

	reqURL := q.formUrl(c.config.PlatformServer, c.config.PlatformAccessToken)

	// Pass context for proper timeout/cancellation handling
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return oculusResp, fmt.Errorf("failed to create ban request: %w", err)
	}

	myCFG.DebugLogPrintf("RequestOculusAttestationBan : %+v", req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return oculusResp, fmt.Errorf("oculus ban api request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return oculusResp, fmt.Errorf("failed to read response body: %w", err)
	}

	myCFG.DebugLogPrintf("RequestOculusAttestationBan resp : %q", respBytes)

	// Handle non-200 responses explicitly
	if resp.StatusCode != http.StatusOK {
		return oculusResp, fmt.Errorf("oculus attestation ban error (status %d): %s", resp.StatusCode, string(respBytes))
	}

	if err := json.Unmarshal(respBytes, &oculusResp); err != nil {
		return oculusResp, fmt.Errorf("failed to unmarshal oculus ban response: %w", err)
	}

	return oculusResp, nil
}

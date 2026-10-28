package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	myCFG "gofiber/config"
	myMDL "gofiber/model"
	"log"
	"net/http"
	"time"
)

type CloudflareIPBanService interface {
	ListBannedIPs(ctx context.Context, q *myMDL.ListQuery) ([]CloudflareListItem, *CloudflareResultInfo, error)
	//BlockIP(ctx context.Context, ip string, reason string) error
	//UnblockIP(ctx context.Context, ip string) error
}

type cloudflareIPBanServiceImpl struct {
	apiToken string
	//accountID string
	//listID    string
	zoneID string
	client *http.Client
}

// func (cf *cloudflareIPBanServiceImpl) formListBannedIPsUrl() string {
// 	return fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/firewall/access_rules/rules", cf.accountID)
// }

func (cf *cloudflareIPBanServiceImpl) formListBannedIPsUrl() string {
	// Replaced /accounts/%s/ with /zones/%s/
	return fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/firewall/access_rules/rules", cf.zoneID)
}

func (cf *cloudflareIPBanServiceImpl) formBlockIPUrl() string {
	return cf.formListBannedIPsUrl()
}

func NewCloudflareIPBanService(apiToken, zoneID string) CloudflareIPBanService {
	return &cloudflareIPBanServiceImpl{
		apiToken: apiToken,
		zoneID:   zoneID,
		client:   &http.Client{Timeout: 5 * time.Second},
	}
}

type CloudflareResultInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
}

type CloudflareConfiguration struct {
	Target string `json:"target"` // e.g., "ip", "ip_range", "asn", "country"
	Value  string `json:"value"`  // The actual IP address or CIDR (e.g., "192.168.1.1")
}

type CloudflareListItem struct {
	ID            string                  `json:"id"`
	Mode          string                  `json:"mode"`
	Notes         string                  `json:"notes"` // Cloudflare uses "notes" instead of "comment"
	CreatedOn     string                  `json:"created_on"`
	Configuration CloudflareConfiguration `json:"configuration"`
}

func (item *CloudflareListItem) IP() string {
	return item.Configuration.Value
}

func (item *CloudflareListItem) Comment() string {
	return item.Notes
}

func (cf *cloudflareIPBanServiceImpl) formListBannedIPsUrlQuery(q *myMDL.ListQuery) string {
	// Replaced /accounts/%s/ with /zones/%s/
	return fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/firewall/access_rules/rules?page=%d&per_page=%d", cf.zoneID, q.Page, q.PagePerCount)
}

func (cf *cloudflareIPBanServiceImpl) formListBannedIPsRequest(ctx context.Context, q *myMDL.ListQuery) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cf.formListBannedIPsUrlQuery(q), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+cf.apiToken)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (cf *cloudflareIPBanServiceImpl) ListBannedIPs(ctx context.Context, q *myMDL.ListQuery) ([]CloudflareListItem, *CloudflareResultInfo, error) {
	if !myCFG.CloudflareEnabled() {
		log.Println("[cloudflare] skip list banned ip")
		return []CloudflareListItem{}, nil, nil
	}

	req, err := cf.formListBannedIPsRequest(ctx, q)
	if err != nil {
		return nil, nil, err
	}

	resp, err := cf.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var cfResponse struct {
		Success              bool                 `json:"success"`
		Result               []CloudflareListItem `json:"result"`
		CloudflareResultInfo `json:"result_info"`
		Errors               []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&cfResponse); err != nil {
		return nil, nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	// 2. Validate HTTP status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("cloudflare API HTTP status %d: %v", resp.StatusCode, cfResponse.Errors)
	}

	// 3. Validate Cloudflare application-level success flag
	if !cfResponse.Success {
		return nil, nil, fmt.Errorf("cloudflare API request failed: %v", cfResponse.Errors)
	}

	return cfResponse.Result, &cfResponse.CloudflareResultInfo, nil
}

func (cf *cloudflareIPBanServiceImpl) formBlockIPRequest(ctx context.Context, payload []map[string]interface{}) (*http.Request, error) {
	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cf.formBlockIPUrl(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+cf.apiToken)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// BlockIP appends an IP to a Cloudflare Bulk Redirect/IP List
func (cf *cloudflareIPBanServiceImpl) BlockIP(ctx context.Context, ip string, reason string) error {
	if !myCFG.CloudflareEnabled() {
		log.Println("[cloudflare] skip block ip")
		return nil
	}

	payload := []map[string]interface{}{
		{
			"ip":      ip,
			"comment": reason,
		},
	}

	req, err := cf.formBlockIPRequest(ctx, payload)
	if err != nil {
		return err
	}

	resp, err := cf.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("cloudflare API error: status %d", resp.StatusCode)
	}

	return nil
}

func (cf *cloudflareIPBanServiceImpl) UnblockIP(ctx context.Context, ip string) error {
	if !myCFG.CloudflareEnabled() {
		log.Println("[cloudflare] skip unblock ip")
		return nil
	}

	// Implement Cloudflare DELETE list item logic here
	return nil
}

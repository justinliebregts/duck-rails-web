package model

import (
	"time"
)

type OCULUSResponseError struct {
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
	Code    int    `json:"code,omitempty"`
	//ErrorDate    []string `json:"error_data,omitempty"`
	ErrorSubcode int    `json:"error_subcode,omitempty"`
	FBTraceID    string `json:"fbtrace_id,omitempty"`
}

type OCULUSResponseBase struct {
	Success bool `json:"success"`
	Error   OCULUSResponseError
}

type VerifyItemOwnershipQuery struct {
	//UserID      string `json:"user_id"`
	SKU   string `json:"sku"`
	UsrID string `json:"user_id"`
	//AccessToken string `json:"access_token"`
}

type RetrieveItemsOwnedQuery struct {
	OrgScopedID string   `json:"user_id"`
	Fields      []string `json:"fields"`
}

type OculusItem struct {
	SKU string `json:"sku"`
	ID  string `json:"id"`
}

type OculusData struct {
	ID             string     `json:"id"`
	GrantTime      int64      `json:"grant_time"`
	ExpirationTime int64      `json:"expiration_time"`
	Item           OculusItem `json:"item"`
}

type OculusCursors struct {
	After  string `json:"after"`
	Before string `json:"before"`
}

type OculusPaging struct {
	Cursors  OculusCursors `json:"cursors"`
	Previous string        `json:"previous"`
	Next     string        `json:"next"`
}

type OculusError struct {
	Message      string            `json:"message"`
	Type         string            `json:"type"`
	Code         int               `json:"code"`
	ErrorData    map[string]string `json:"error_data"`
	ErrorSubcode int               `json:"error_subcode"`
	FbTraceID    string            `json:"fbtrace_id"`
}

type RetrieveItemsOwnedResponse struct {
	Data   []OculusData `json:"data"`
	Paging OculusPaging `json:"paging"`
	Error  OculusError  `json:"error,omitempty"`
}

//type ConsumeIAPItemQuery struct {
//	SKU         string `json:"sku"`
//	AccessToken string `json:"access_token"`
//}
//
//func (c *ConsumeIAPItemQuery) BuildQuery() url.Values {
//	params := url.Values{}
//	params.Add("sku", c.SKU)
//	params.Add("access_token", c.AccessToken)
//
//	log.Println(params)
//	return params
//}

type OculusConsumeIAPItemQuery struct {
	VerifyItemOwnershipQuery
}

type UserNonceValidateQuery struct {
	UserID string `json:"user_id"`
	Nonce  string `json:"nonce"`

	RequestTimeout time.Duration `json:"timeout"`
}

type UserNonceValidateResponse struct {
	IsValid bool                `json:"is_valid"`
	Error   OCULUSResponseError `json:"error"`
}

type GetOculusOrgScopedIDResponseQuery struct {
	Fields []string `json:"fields"`
	//AccessToken string   `json:"access_token"`
}

type GetOculusOrgScopedIDResponse struct {
	ID       string `json:"id"`
	Alias    string `json:"alias"`
	ScopedID string `json:"org_scoped_id"`
}

func (r *GetOculusOrgScopedIDResponse) IsValid() bool {
	if len(r.ScopedID) <= 0 || len(r.ID) <= 0 {
		return false
	}
	return true
}

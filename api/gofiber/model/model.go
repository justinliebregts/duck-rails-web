package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

const (
	ThreeSeconds = time.Second * 3

	DateFormat     = "2006-01-02"
	DatetimeFormat = "2006-01-02 15:04:05"
)

var (
	gmtMinus7 = time.FixedZone("GMT-7", -7*60*60)

	// use a single instance of Validate, it caches struct info
	validate       = validator.New()
	Default200Resp = ResponseGeneral{Status: fiber.StatusOK, Success: true}
	Default201Resp = ResponseGeneral{Status: fiber.StatusCreated, Success: true}

	Default400Resp = ResponseGeneral{Status: fiber.StatusBadRequest, Error: "BadRequest400"}
	Default401Resp = ResponseGeneral{Status: fiber.StatusUnauthorized, Error: "Unauthorized"}
	Default402Resp = ResponseGeneral{Status: fiber.StatusPaymentRequired, Error: "PaymentRequired"}
	Default403Resp = ResponseGeneral{Status: fiber.StatusForbidden, Error: "Forbidden"}
	Default404Resp = ResponseGeneral{Status: fiber.StatusNotFound, Error: "Not Found"}
	Default405Resp = ResponseGeneral{Status: fiber.StatusMethodNotAllowed, Error: "Method Not Allowed"}
	Default409Resp = ResponseGeneral{Status: fiber.StatusConflict, Error: "Conflict"}

	Default429Resp = ResponseGeneral{Status: fiber.StatusTooManyRequests, Error: "Too Many Requests", Data: ""}

	Default500Resp = ResponseGeneral{Status: fiber.StatusInternalServerError, Error: "Internal Server Error"}

	Default503Resp = ResponseGeneral{Status: fiber.StatusServiceUnavailable, Error: "Service Unavailable"}
)

type ResponseGeneral struct {
	Status   int         `json:"status"`
	Data     interface{} `json:"data,omitempty"`
	Meta     interface{} `json:"meta,omitempty"`
	Message  string      `json:"message,omitempty"`
	Messages []string    `json:"messages,omitempty"`
	Error    string      `json:"error,omitempty"`
	Success  bool        `json:"success"`
}

func (r *ResponseGeneral) Copy() (resp ResponseGeneral) {

	resp.Status = r.Status
	resp.Data = r.Data
	resp.Message = r.Message
	resp.Error = r.Error
	resp.Success = r.Success

	if r.Messages != nil {
		resp.Messages = make([]string, len(r.Messages))
		copy(resp.Messages, r.Messages)
	} else {
		resp.Messages = nil
	}
	return
}

func (r *ResponseGeneral) SetErrorAndMessage(txt string) {
	r.Message = txt
	r.AddMessages(txt)
	r.Error = txt
}

func (r *ResponseGeneral) AddErrorAsMessages(errs ...error) {
	msgs := make([]string, len(errs))
	for idx := range errs {
		msgs[idx] = errs[idx].Error()
	}

	r.AddMessages(msgs...)
}

func (r *ResponseGeneral) AddDateAsMessages(tms ...time.Time) {
	msgs := make([]string, len(tms))
	for idx := range tms {
		msgs[idx] = tms[idx].Format(DateFormat)
	}

	r.AddMessages(msgs...)
}

func (r *ResponseGeneral) AddTimesAsMessages(tms ...time.Time) {
	msgs := make([]string, len(tms))
	for idx := range tms {
		msgs[idx] = tms[idx].Format(DatetimeFormat)
	}

	r.AddMessages(msgs...)
}

func (r *ResponseGeneral) AddMessages(msgs ...string) {
	r.Messages = append(r.Messages, msgs...)
}

type JsonNullTime struct {
	sql.NullTime
}

func (v JsonNullTime) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Time)
	} else {
		return json.Marshal(nil)
	}
}

func (v *JsonNullTime) UnmarshalJSON(data []byte) error {
	var x *time.Time
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Time = *x
	} else {
		v.Valid = false
	}
	return nil
}

type JsonTime struct {
	time.Time
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (jt *JsonTime) Scan(value interface{}) error {
	if value == nil {
		jt.Time = time.Time{}
		return nil
	}

	t, ok := value.(time.Time)
	if !ok {
		return errors.New("failed to scan JsonTime value")
	}

	jt.Time = t
	return nil
}

// Value return json value, implement driver.Valuer interface
func (jt JsonTime) Value() (driver.Value, error) {
	return jt.MarshalJSON()
}

func (jt JsonTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + jt.Time.Format(DatetimeFormat) + `"`), nil
}

func (jt *JsonTime) UnmarshalJSON(data []byte) error {
	// Remove double quotes around the input string
	strData := strings.Trim(string(data), `"`)

	// Parse the input string as a time value in the desired format
	if t, err := time.Parse(DatetimeFormat, strData); err != nil {
		return err
	} else {
		// Assign the resulting time value to the JsonTime receiver
		jt.Time = t
	}
	return nil
}

func CurrentTimestamp() time.Time {
	return time.Now().UTC()
}

func ParseDateString(dateString string) (t time.Time, err error) {
	return time.Parse(time.RFC3339, dateString+"T00:00:00Z")
}

func ParseDateStringLastSecond(dateString string) (t time.Time, err error) {
	return time.Parse(time.RFC3339, dateString+"T23:59:59Z")
}

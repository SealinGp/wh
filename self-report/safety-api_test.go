package self_report

import (
	"bytes"
	"net/http"
	"testing"
)

const (
	url          = "http://localhost:1234/report"
	content_type = "application/json"
)

//
func TestSafetyApi_SaveOrUpdate(t *testing.T) {
	body := `{
    "initAreaCode": "",
    "randCode_": "",
    "innerId": "700EF8026D9AE6D6E0530100007F4FE8",
    "unitType": "",
    "unitId": "",
    "userId": "",
    "selfcheckDefineID": "7595255C359F7018E0530100007FE0E2",
    "entpTypeId": "",
    "isMajor": "",
    "cookie-name": "3B4A770C3AF55A22884CD9C5F462DF3E",
    "cookie-value": "700EF8026D9CE6D6E0530100007F4FE81618799806154",
    "cookie-path": "",
    "cookie-domain": "",
    "isCusotm": "",
    "reportFailedNum": 0
}`

	resp, err := http.Post(url, content_type, bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatalf("[E] %v", err)
		return
	}
	t.Logf("[I] resp:%v", resp)
}

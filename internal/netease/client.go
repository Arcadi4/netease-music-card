package netease

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	userID     string
	cookie     string
}

func NewClient(userID, cookie string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userID:     userID,
		cookie:     cookie,
	}
}

func (c *Client) request(apiURL string, params map[string]interface{}) (map[string]interface{}, error) {
	encrypted, err := EncryptWeapi(params)
	if err != nil {
		return nil, fmt.Errorf("encrypt params: %w", err)
	}

	formData := url.Values{}
	formData.Set("params", encrypted["params"])
	formData.Set("encSecKey", encrypted["encSecKey"])

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", fmt.Sprintf("MUSIC_U=%s", c.cookie))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	code, ok := result["code"].(float64)
	if !ok || code != 200 {
		if code == 301 || code == 400 || code == 401 {
			return nil, fmt.Errorf("Cookie appears invalid or expired")
		}
		return nil, fmt.Errorf("API error: code=%v", result["code"])
	}

	if account, exists := result["account"]; exists {
		if account == nil {
			return nil, fmt.Errorf("Cookie appears invalid or expired")
		}
	}

	return result, nil
}

func (c *Client) UserAccount() (map[string]interface{}, error) {
	return c.request("https://music.163.com/weapi/nuser/account/get", map[string]interface{}{})
}

func (c *Client) UserRecord(uid string, recordType int) (map[string]interface{}, error) {
	return c.request("https://music.163.com/weapi/v1/play/record", map[string]interface{}{
		"uid":  uid,
		"type": recordType,
	})
}

func (c *Client) SongDetail(ids []string) (map[string]interface{}, error) {
	idsJSON, _ := json.Marshal(ids)
	return c.request("https://music.163.com/weapi/v3/song/detail", map[string]interface{}{
		"c":   fmt.Sprintf("[%s]", strings.Join(ids, ",")),
		"ids": string(idsJSON),
	})
}

func (c *Client) UserDetail(uid string) (map[string]interface{}, error) {
	return c.request("https://music.163.com/weapi/v1/user/detail/"+uid, map[string]interface{}{})
}

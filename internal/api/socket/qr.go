package socket

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tranhaonguyendev/za-go/internal/util"
)

func applyHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

func doReq(client *http.Client, method string, rawURL string, headers map[string]string, form url.Values) (map[string]any, error) {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequest(method, rawURL, body)
	if err != nil {
		return nil, err
	}
	applyHeaders(req, headers)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	out := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func mustClient(qr *QRAuthResult) (*http.Client, error) {
	if qr == nil {
		return nil, fmt.Errorf("qr auth result is nil")
	}
	if qr.client == nil {
		return nil, fmt.Errorf("qr auth session is unavailable")
	}
	return qr.client, nil
}

func qrHeaders(referer string) map[string]string {
	return map[string]string{
		"accept":          "*/*",
		"accept-language": "vi,en-US;q=0.9,en;q=0.8",
		"content-type":    "application/x-www-form-urlencoded",
		"origin":          "https://id.zalo.me",
		"referer":         referer,
		"user-agent":      util.HEADERS["User-Agent"],
	}
}

func userInfoHeaders() map[string]string {
	return map[string]string{
		"accept":          "*/*",
		"accept-language": "vi-VN,vi;q=0.9,en-US;q=0.6,en;q=0.5",
		"referer":         "https://chat.zalo.me/",
		"user-agent":      util.HEADERS["User-Agent"],
	}
}

func (s *SocketAPI) AuthQRCode() (*QRAuthResult, error) {
	client := freshClient()
	headers := map[string]string{
		"User-Agent":      util.HEADERS["User-Agent"],
		"Accept":          "application/json, text/plain, */*",
		"Origin":          "https://chat.zalo.me",
		"Referer":         "https://chat.zalo.me/",
		"Accept-Language": "vi-VN,vi;q=0.9,en-US;q=0.6,en;q=0.5",
	}
	_, _ = client.Get("https://id.zalo.me/account?continue=https%3A%2F%2Fchat.zalo.me%2F")
	_, _ = doReq(client, http.MethodPost, "https://id.zalo.me/account/logininfo", headers, url.Values{"continue": {"https://chat.zalo.me/"}, "v": {"5.5.7"}})
	_, _ = doReq(client, http.MethodPost, "https://id.zalo.me/account/verify-client", map[string]string{"User-Agent": util.HEADERS["User-Agent"], "Origin": "https://id.zalo.me", "Referer": "https://id.zalo.me/account?continue=https%3A%2F%2Fchat.zalo.me%2F", "Accept": "*/*"}, url.Values{"type": {"device"}, "continue": {"https://zalo.me/pc"}, "v": {"5.5.7"}})
	data, err := doReq(client, http.MethodPost, "https://id.zalo.me/account/authen/qr/generate", map[string]string{"User-Agent": util.HEADERS["User-Agent"], "Origin": "https://id.zalo.me", "Referer": "https://id.zalo.me/account?continue=https%3A%2F%2Fchat.zalo.me%2F", "Accept": "*/*"}, url.Values{"continue": {"https://zalo.me/pc"}, "v": {"5.5.7"}})
	if err != nil {
		return nil, err
	}
	payload := util.AsMap(data["data"])
	image := util.AsString(payload["image"])
	code := util.AsString(payload["code"])
	token := util.AsString(payload["token"])
	if image == "" || code == "" {
		return nil, fmt.Errorf("unable to generate qr code")
	}
	image = strings.TrimPrefix(image, "data:image/png;base64,")
	imageBytes, err := base64.StdEncoding.DecodeString(image)
	if err != nil {
		return nil, err
	}
	imagePath := filepath.Join("assets", "attachments", "authQR.png")
	if err := os.MkdirAll(filepath.Dir(imagePath), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(imagePath, imageBytes, 0o644); err != nil {
		return nil, err
	}
	cookies := qrCookieMap(client.Jar)
	return &QRAuthResult{
		Code:       code,
		Token:      token,
		ImageBytes: imageBytes,
		ImagePath:  imagePath,
		Cookies:    cookies,
		client:     client,
	}, nil
}

func (s *SocketAPI) WaitQRCodeScan(qr *QRAuthResult, maxAttempts int, interval time.Duration) (bool, error) {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	if interval <= 0 {
		interval = 3 * time.Second
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		data, err := s.CheckQRCodeScan(qr)
		if err != nil {
			return false, err
		}
		switch util.AsInt(data["error_code"]) {
		case 0:
			return true, nil
		case 1:
			if attempt < maxAttempts {
				time.Sleep(interval)
			}
		default:
			return false, fmt.Errorf("waiting-scan failed: %#v", data)
		}
	}
	return false, nil
}

func (s *SocketAPI) CheckQRCodeScan(qr *QRAuthResult) (map[string]any, error) {
	client, err := mustClient(qr)
	if err != nil {
		return nil, err
	}
	payload := url.Values{
		"code":     {qr.Code},
		"continue": {"https://chat.zalo.me/"},
		"v":        {"5.5.7"},
	}
	headers := qrHeaders("https://id.zalo.me/account?continue=https%3A%2F%2Fchat.zalo.me%2F")
	resp, err := doReq(client, http.MethodPost, "https://id.zalo.me/account/authen/qr/waiting-scan", headers, payload)
	if err == nil {
	}
	return resp, err
}

func (s *SocketAPI) WaitQRCodeConfirm(qr *QRAuthResult, maxAttempts int, interval time.Duration) (map[string]string, error) {
	if maxAttempts <= 0 {
		maxAttempts = 10
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		data, err := s.CheckQRCodeConfirm(qr)
		if err != nil {
			return nil, err
		}
		if util.AsInt(data["error_code"]) == 0 {
			client, err := mustClient(qr)
			if err != nil {
				return nil, err
			}

			cookies := qrCookieMap(client.Jar)
			qr.Cookies = cookies

			_, _ = s.CheckQRSession(qr)
			cookies = qrCookieMap(client.Jar)
			qr.Cookies = cookies

			_, _ = s.FetchQRUserInfo(qr)
			cookies = qrCookieMap(client.Jar)
			qr.Cookies = cookies

			return cookies, nil
		}
		if attempt < maxAttempts {
			time.Sleep(interval)
		}
	}
	return nil, fmt.Errorf("waiting-confirm timed out for qr code %s", qr.Code)
}

func (s *SocketAPI) CheckQRCodeConfirm(qr *QRAuthResult) (map[string]any, error) {
	client, err := mustClient(qr)
	if err != nil {
		return nil, err
	}
	payload := url.Values{
		"code":     {qr.Code},
		"gToken":   {""},
		"gAction":  {"CONFIRM_QR"},
		"continue": {"https://chat.zalo.me/index.html"},
		"v":        {"5.5.7"},
	}
	headers := qrHeaders("https://id.zalo.me/account?continue=https%3A%2F%2Fchat.zalo.me%2F")
	data, err := doReq(client, http.MethodPost, "https://id.zalo.me/account/authen/qr/waiting-confirm", headers, payload)
	if err != nil {
		return nil, err
	}
	if util.AsInt(data["error_code"]) == 0 {
		qr.Cookies = qrCookieMap(client.Jar)
	}
	return data, nil
}

func (s *SocketAPI) CheckQRSession(qr *QRAuthResult) (string, error) {
	client, err := mustClient(qr)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodGet, "https://id.zalo.me/account/checksession?continue=https%3A%2F%2Fchat.zalo.me%2Findex.html", nil)
	if err != nil {
		return "", err
	}
	applyHeaders(req, map[string]string{
		"accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"referer":    "https://id.zalo.me/account?continue=https%3A%2F%2Fchat.zalo.me%2F",
		"user-agent": util.HEADERS["User-Agent"],
	})
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (s *SocketAPI) FetchQRUserInfo(qr *QRAuthResult) (map[string]any, error) {
	client, err := mustClient(qr)
	if err != nil {
		return nil, err
	}
	resp, err := doReq(client, http.MethodGet, "https://jr.chat.zalo.me/jr/userinfo", userInfoHeaders(), nil)
	if err == nil {
	}
	return resp, err
}

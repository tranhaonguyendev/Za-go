package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nguyendev/zago/internal/util"
)

type MultipartFile struct {
	FieldName   string
	FileName    string
	Content     []byte
	ContentType string
}

type State struct {
	Config       map[string]any
	Headers      map[string]string
	Cookies      map[string]string
	Session      *http.Client
	NoRedirect   *http.Client
	UserClientID string
	ClientUUID   string
	Loggin       bool
	PhoneNumber  string
	Password     string
	ZpwWs        string
}

func NewState() *State {
	jar, _ := cookiejar.New(nil)
	headers := map[string]string{}
	for k, v := range util.HEADERS {
		headers[k] = v
	}
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:   false,
		MaxIdleConns:        128,
		MaxIdleConnsPerHost: 32,
		MaxConnsPerHost:     32,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 15 * time.Second,
		DisableCompression:  true,
	}
	return &State{
		Config:  map[string]any{},
		Headers: headers,
		Cookies: map[string]string{},
		Session: &http.Client{Jar: jar, Transport: transport},
		NoRedirect: &http.Client{
			Jar:       jar,
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (s *State) IsLoggedin() bool              { return s.Loggin }
func (s *State) GetCookies() map[string]string { return s.Cookies }

func (s *State) SetCookies(cookies map[string]string) {
	if cookies == nil {
		cookies = map[string]string{}
	}
	s.Cookies = cookies
	if s.Session != nil && s.Session.Jar != nil {
		hosts := []string{"https://chat.zalo.me", "https://wpa.chat.zalo.me", "https://tt-group-wpa.chat.zalo.me", "https://tt-files-wpa.chat.zalo.me", "https://id.zalo.me"}
		for _, raw := range hosts {
			u, _ := url.Parse(raw)
			jarCookies := make([]*http.Cookie, 0, len(cookies))
			for k, v := range cookies {
				if strings.TrimSpace(k) == "" {
					continue
				}
				jarCookies = append(jarCookies, &http.Cookie{Name: k, Value: v})
			}
			s.Session.Jar.SetCookies(u, jarCookies)
		}
	}
}

func (s *State) GetSecretkey() string {
	keys := []string{"secretkey", "secret_key", "zpw_enk", "zpwEnk"}
	for _, k := range keys {
		if v, ok := s.Config[k]; ok {
			if vs := strings.TrimSpace(fmt.Sprintf("%v", v)); vs != "" {
				return vs
			}
		}
	}
	return ""
}

func (s *State) SetSecretkey(secretkey string) {
	s.Config["secretkey"] = secretkey
	s.Config["secret_key"] = secretkey
	s.Config["zpw_enk"] = secretkey
	s.Config["zpwEnk"] = secretkey
}

func (s *State) SetPhoneNumber(phone string) {
	s.PhoneNumber = phone
	s.Config["phone_number"] = phone
	s.Config["phoneNumber"] = phone
}

func (s *State) SetPassword(password string, storeInConfig bool) {
	s.Password = password
	if storeInConfig {
		s.Config["password"] = password
	}
}

func (s *State) applyUserAgent(userAgent string) {
	if strings.TrimSpace(userAgent) != "" {
		s.Headers["User-Agent"] = userAgent
	}
}

func (s *State) addHeaders(req *http.Request, extra map[string]string) {
	for k, v := range s.Headers {
		req.Header.Set(k, v)
	}
	for k, v := range extra {
		req.Header.Set(k, v)
	}
	if raw := util.DictToRawCookies(s.Cookies); raw != "" {
		req.Header.Set("Cookie", raw)
	}
}

func (s *State) buildURL(rawURL string, query url.Values) string {
	if len(query) == 0 {
		return rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	for k, vals := range query {
		for _, v := range vals {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *State) do(req *http.Request, timeout time.Duration, allowRedirects bool) (*http.Response, error) {
	client := s.Session
	if !allowRedirects && s.NoRedirect != nil {
		client = s.NoRedirect
	}
	if timeout > 0 {
		ctx, cancel := context.WithTimeout(req.Context(), timeout)
		defer cancel()
		req = req.WithContext(ctx)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(b))
	}
	return resp, nil
}

func (s *State) GetSession(rawURL string) (*http.Response, error) {
	return s.GetSessionWithParams(rawURL, nil)
}

func (s *State) GetSessionWithParams(rawURL string, query url.Values) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, s.buildURL(rawURL, query), nil)
	if err != nil {
		return nil, err
	}
	s.addHeaders(req, nil)
	return s.do(req, 0, true)
}

func (s *State) GetSessionEx(rawURL string, query url.Values, timeout time.Duration, allowRedirects bool, extraHeaders map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, s.buildURL(rawURL, query), nil)
	if err != nil {
		return nil, err
	}
	s.addHeaders(req, extraHeaders)
	return s.do(req, timeout, allowRedirects)
}

func (s *State) HeadSession(rawURL string, timeout time.Duration) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodHead, rawURL, nil)
	if err != nil {
		return nil, err
	}
	s.addHeaders(req, nil)
	return s.do(req, timeout, true)
}

func (s *State) PostSession(rawURL string, form url.Values) (*http.Response, error) {
	return s.PostSessionWithParams(rawURL, nil, form)
}

func (s *State) PostSessionWithParams(rawURL string, query url.Values, form url.Values) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, s.buildURL(rawURL, query), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	s.addHeaders(req, map[string]string{"Content-Type": "application/x-www-form-urlencoded"})
	return s.do(req, 0, true)
}

func (s *State) PostSessionBody(rawURL string, query url.Values, body io.Reader, contentType string, timeout time.Duration, extraHeaders map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, s.buildURL(rawURL, query), body)
	if err != nil {
		return nil, err
	}
	h := map[string]string{}
	for k, v := range extraHeaders {
		h[k] = v
	}
	if contentType != "" {
		h["Content-Type"] = contentType
	}
	s.addHeaders(req, h)
	return s.do(req, timeout, true)
}

func (s *State) PostMultipartSession(rawURL string, query url.Values, fields map[string]string, files []MultipartFile, timeout time.Duration) (*http.Response, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	for _, file := range files {
		part, err := w.CreateFormFile(file.FieldName, file.FileName)
		if err != nil {
			return nil, err
		}
		if _, err := part.Write(file.Content); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return s.PostSessionBody(rawURL, query, &buf, w.FormDataContentType(), timeout, nil)
}

func (s *State) buildLoginInfoURL(imei string) string {
	q := url.Values{}
	q.Set("imei", imei)
	q.Set("type", "30")
	q.Set("client_version", "645")
	q.Set("computer_name", "Web")
	q.Set("ts", strconv.FormatInt(util.Now(), 10))
	return "https://wpa.chat.zalo.me/api/login/getLoginInfo?" + q.Encode()
}

func firstValue(data map[string]any, keys ...string) any {
	for _, k := range keys {
		v, ok := data[k]
		if !ok || v == nil {
			continue
		}
		if s, ok := v.(string); ok {
			s = strings.TrimSpace(s)
			if s == "" || s == "<nil>" {
				continue
			}
			return s
		}
		return v
	}
	return nil
}

func (s *State) mergeLoginData(data map[string]any, imei string) error {
	errCode, _ := data["error_code"].(float64)
	errMsg, _ := data["error_message"].(string)
	payload, _ := data["data"].(map[string]any)
	if int(errCode) != 0 || payload == nil {
		if errMsg == "" {
			errMsg = "Undefined error"
		}
		return fmt.Errorf("error #%d during login: %s", int(errCode), errMsg)
	}
	secret := firstString(payload, "zpw_enk", "zpwEnk", "secret_key", "secretkey", "secretKey")
	if secret == "" {
		return fmt.Errorf("unable to retrieve secret key")
	}
	s.Config = map[string]any{}
	for k, v := range payload {
		s.Config[k] = v
	}
	s.SetSecretkey(secret)
	phone := firstString(payload, "phone_number", "phoneNumber", "phone")
	uid := firstString(payload, "send2me_id", "send2meId", "uid")
	zpwWsRaw := firstValue(payload, "zpw_ws", "zpwWs", "zpw_ws_v2", "zpwWsV2")
	zpwWs := ""
	switch v := zpwWsRaw.(type) {
	case []any:
		for _, item := range v {
			if candidate := strings.TrimSpace(fmt.Sprintf("%v", item)); candidate != "" && candidate != "<nil>" {
				zpwWs = candidate
				break
			}
		}
	case []string:
		for _, item := range v {
			if candidate := strings.TrimSpace(item); candidate != "" {
				zpwWs = candidate
				break
			}
		}
	default:
		zpwWs = strings.TrimSpace(fmt.Sprintf("%v", zpwWsRaw))
	}
	s.SetPhoneNumber(phone)
	s.Config["send2me_id"] = uid
	s.Config["send2meId"] = uid
	if zpwWsRaw != nil {
		s.Config["zpw_ws"] = zpwWsRaw
		s.Config["zpwWs"] = zpwWsRaw
	} else {
		s.Config["zpw_ws"] = zpwWs
		s.Config["zpwWs"] = zpwWs
	}
	s.UserClientID = uid
	s.ClientUUID = imei
	s.ZpwWs = zpwWs
	s.Loggin = true
	return nil
}

func firstString(data map[string]any, keys ...string) string {
	for _, k := range keys {
		v, ok := data[k]
		if !ok {
			continue
		}
		s := strings.TrimSpace(fmt.Sprintf("%v", v))
		if s != "" && s != "<nil>" {
			return s
		}
	}
	return ""
}

func (s *State) tryCookieLogin(imei string) error {
	if len(s.Cookies) == 0 {
		return fmt.Errorf("login method not supported")
	}
	resp, err := s.GetSession(s.buildLoginInfoURL(imei))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var parsed map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return err
	}
	return s.mergeLoginData(parsed, imei)
}

func (s *State) hydrateFromConfig(imei string) {
	s.Loggin = true
	if s.ClientUUID == "" {
		s.ClientUUID = imei
	}
	if s.UserClientID == "" {
		s.UserClientID = firstString(s.Config, "send2me_id", "send2meId", "uid")
	}
	if s.PhoneNumber == "" {
		s.PhoneNumber = firstString(s.Config, "phone_number", "phoneNumber", "phone")
	}
	if s.ZpwWs == "" {
		s.ZpwWs = firstString(s.Config, "zpw_ws", "zpwWs", "zpw_ws_v2", "zpwWsV2")
	}
}

func (s *State) Login(phone, password, imei string, sessionCookies map[string]string, userAgent string) error {
	if strings.TrimSpace(password) != "" {
		s.SetPassword(password, true)
	}
	if strings.TrimSpace(phone) != "" {
		s.SetPhoneNumber(phone)
	}
	s.applyUserAgent(userAgent)
	if sessionCookies != nil {
		s.SetCookies(sessionCookies)
	}
	if len(s.Cookies) > 0 && s.GetSecretkey() != "" {
		s.hydrateFromConfig(imei)
		return nil
	}
	if err := s.tryCookieLogin(imei); err != nil {
		return err
	}
	return nil
}

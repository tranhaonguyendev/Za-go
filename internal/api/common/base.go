package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tranhaonguyendev/za-go/internal/app"
	core "github.com/tranhaonguyendev/za-go/internal/core"
	"github.com/tranhaonguyendev/za-go/internal/util"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

type BaseAPI struct {
	State        *app.State
	APILoginType int
	IMEI         string
	UID          string
	Language     string
	Hub          *worker.Hub
}

func NewBaseAPI(state *app.State, loginType int, hub *worker.Hub) *BaseAPI {
	return &BaseAPI{State: state, APILoginType: loginType, Language: "vi", Hub: hub}
}

func (b *BaseAPI) IMEIValue() string {
	if strings.TrimSpace(b.IMEI) != "" {
		return b.IMEI
	}
	if b.State != nil {
		return b.State.ClientUUID
	}
	return ""
}

func (b *BaseAPI) UIDValue() string {
	if strings.TrimSpace(b.UID) != "" {
		return b.UID
	}
	if b.State != nil {
		return b.State.UserClientID
	}
	return ""
}

func (b *BaseAPI) Secret() string {
	if b.State == nil {
		return ""
	}
	return b.State.GetSecretkey()
}

func (b *BaseAPI) Query(params map[string]any) url.Values {
	q := url.Values{}
	for k, v := range params {
		switch t := v.(type) {
		case string:
			q.Set(k, t)
		case int:
			q.Set(k, strconv.Itoa(t))
		case int64:
			q.Set(k, strconv.FormatInt(t, 10))
		case float64:
			q.Set(k, strconv.FormatInt(int64(t), 10))
		case bool:
			if t {
				q.Set(k, "1")
			} else {
				q.Set(k, "0")
			}
		default:
			q.Set(k, fmt.Sprintf("%v", v))
		}
	}
	return q
}

func (b *BaseAPI) Encode(params any) (string, error) {
	return util.ZaloEncode(params, b.Secret())
}

func (b *BaseAPI) EncodedForm(payload any) (url.Values, error) {
	enc, err := b.Encode(payload)
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("params", enc)
	return form, nil
}

func (b *BaseAPI) GetJSON(rawURL string, query url.Values) (map[string]any, error) {
	resp, err := b.State.GetSessionWithParams(rawURL, query)
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

func (b *BaseAPI) GetJSONEx(rawURL string, query url.Values, timeout time.Duration, allowRedirects bool) (map[string]any, error) {
	resp, err := b.State.GetSessionEx(rawURL, query, timeout, allowRedirects, nil)
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

func (b *BaseAPI) PostJSON(rawURL string, query url.Values, form url.Values) (map[string]any, error) {
	resp, err := b.State.PostSessionWithParams(rawURL, query, form)
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

func (b *BaseAPI) PostMultipartJSON(rawURL string, query url.Values, fields map[string]string, files []app.MultipartFile, timeout time.Duration) (map[string]any, error) {
	resp, err := b.State.PostMultipartSession(rawURL, query, fields, files, timeout)
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

func (b *BaseAPI) PostBodyJSON(rawURL string, query url.Values, body io.Reader, contentType string, timeout time.Duration) (map[string]any, error) {
	resp, err := b.State.PostSessionBody(rawURL, query, body, contentType, timeout, nil)
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

func (b *BaseAPI) ParseEnvelope(resp map[string]any) (any, error) {
	return util.ParseResponseEnvelope(resp, b.Secret())
}

func (b *BaseAPI) ParseThread(resp map[string]any, threadType core.ThreadType) (any, error) {
	decoded, err := b.ParseEnvelope(resp)
	if err != nil {
		return nil, err
	}
	if threadType == core.GROUP {
		return worker.GroupFromDict(decoded), nil
	}
	return worker.UserFromDict(decoded), nil
}

func (b *BaseAPI) ParseUser(resp map[string]any) (*worker.User, error) {
	decoded, err := b.ParseEnvelope(resp)
	if err != nil {
		return nil, err
	}
	return worker.UserFromDict(decoded), nil
}

func (b *BaseAPI) ParseGroup(resp map[string]any) (*worker.Group, error) {
	decoded, err := b.ParseEnvelope(resp)
	if err != nil {
		return nil, err
	}
	return worker.GroupFromDict(decoded), nil
}

func (b *BaseAPI) ParseRaw(resp map[string]any) (any, error) {
	return b.ParseEnvelope(resp)
}

func (b *BaseAPI) ParseStd(resp map[string]any, threadType core.ThreadType, clientID any) (any, error) {
	decoded, err := b.ParseEnvelope(resp)
	if err != nil {
		return nil, err
	}
	if m, ok := decoded.(map[string]any); ok && clientID != nil {
		m["clientId"] = clientID
		decoded = m
		if threadType == core.GROUP {
			return worker.GroupFromDict(decoded), nil
		}
		return worker.UserFromDict(decoded), nil
	}
	if threadType == core.GROUP {
		return worker.GroupFromDict(decoded), nil
	}
	return worker.UserFromDict(decoded), nil
}

func (b *BaseAPI) NormalizeMaybeJSON(v any) any {
	v = util.NormalizeDecodedData(v)
	if s, ok := v.(string); ok {
		var parsed any
		if err := json.Unmarshal([]byte(s), &parsed); err == nil {
			return parsed
		}
	}
	return v
}

func (b *BaseAPI) NormThreadID(threadID string) string {
	threadID = strings.TrimSpace(threadID)
	if threadID == "" {
		return b.UIDValue()
	}
	if n, err := strconv.ParseInt(threadID, 10, 64); err == nil {
		return strconv.FormatInt(n, 10)
	}
	return threadID
}

func (b *BaseAPI) IsURL(v string) bool {
	s := strings.ToLower(strings.TrimSpace(v))
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func (b *BaseAPI) GetExt(path string) string {
	s := strings.Split(strings.Split(path, "?")[0], "#")[0]
	ext := filepath.Ext(s)
	return strings.TrimPrefix(strings.ToLower(ext), ".")
}

func (b *BaseAPI) GetFileName(path string, fallback string) string {
	base := filepath.Base(strings.Split(strings.Split(path, "?")[0], "#")[0])
	if base == "." || base == "/" || base == "" {
		return fallback
	}
	return base
}

func (b *BaseAPI) GetLocalSize(path string) int64 {
	st, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return st.Size()
}

func (b *BaseAPI) RemoteBytes(rawURL string, timeout time.Duration) ([]byte, error) {
	resp, err := b.State.GetSessionEx(rawURL, nil, timeout, true, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (b *BaseAPI) RemoteText(rawURL string, timeout time.Duration) (string, string, error) {
	resp, err := b.State.GetSessionEx(rawURL, nil, timeout, true, nil)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	return string(body), resp.Request.URL.String(), nil
}

func (b *BaseAPI) RemoteHeadSize(rawURL string, timeout time.Duration) int64 {
	resp, err := b.State.HeadSession(rawURL, timeout)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	return resp.ContentLength
}

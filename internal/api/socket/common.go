package socket

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	base "github.com/nguyendev/zago/internal/api/common"
	"github.com/nguyendev/zago/internal/app"
	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

type QRAuthResult struct {
	Code       string
	Token      string
	ImageBytes []byte
	ImagePath  string
	Cookies    map[string]string
	client     *http.Client
}

type SocketAPI struct {
	*base.BaseAPI
	mu          sync.Mutex
	conn        *websocket.Conn
	wsKey       string
	listening   bool
	uploadOnly  bool
	thread      bool
	stopCh      chan struct{}
	pingStopCh  chan struct{}
	dispatchSem chan struct{}
}

func NewSocketAPI(state *app.State, loginType int, hub *worker.Hub) *SocketAPI {
	return &SocketAPI{BaseAPI: base.NewBaseAPI(state, loginType, hub)}
}

func (s *SocketAPI) Listening() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listening
}

func (s *SocketAPI) Ready() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listening && s.conn != nil && strings.TrimSpace(s.wsKey) != ""
}

func (s *SocketAPI) setListening(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listening = v
}

func (s *SocketAPI) SetUploadOnly(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.uploadOnly = v
}

func (s *SocketAPI) UploadOnly() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.uploadOnly
}

func (s *SocketAPI) wsURL() (string, error) {
	urls, err := s.wsURLs()
	if err != nil {
		return "", err
	}
	return urls[0], nil
}

func (s *SocketAPI) wsURLs() ([]string, error) {
	var rawCandidates []string
	switch v := firstNonNil(s.State.Config["zpw_ws"], s.State.Config["zpwWs"], s.State.ZpwWs).(type) {
	case []any:
		for _, item := range v {
			if candidate := util.AsString(item); candidate != "" {
				rawCandidates = append(rawCandidates, candidate)
			}
		}
	case []string:
		for _, item := range v {
			if candidate := strings.TrimSpace(item); candidate != "" {
				rawCandidates = append(rawCandidates, candidate)
			}
		}
	default:
		raw := util.AsString(v)
		if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
			parts := strings.Fields(strings.Trim(raw, "[]"))
			rawCandidates = append(rawCandidates, parts...)
		} else if raw != "" {
			rawCandidates = append(rawCandidates, raw)
		}
	}
	if len(rawCandidates) > 0 {
		urls := make([]string, 0, len(rawCandidates))
		q := url.Values{}
		q.Set("zpw_ver", "647")
		q.Set("zpw_type", fmt.Sprintf("%d", s.APILoginType))
		q.Set("t", fmt.Sprintf("%d", util.Now()))
		for _, raw := range rawCandidates {
			if strings.Contains(raw, "?") {
				urls = append(urls, raw+"&"+q.Encode())
				continue
			}
			urls = append(urls, raw+"?"+q.Encode())
		}
		return urls, nil
	}
	return nil, fmt.Errorf("unable to load websocket url")
}

func (s *SocketAPI) wsHeaders(rawURL string) (http.Header, error) {
	rawCookies := util.DictToRawCookies(s.State.Cookies)
	if rawCookies == "" {
		return nil, fmt.Errorf("unable to load cookies")
	}
	h := http.Header{}
	h.Set("Accept-Language", "en-US,en;q=0.9")
	h.Set("Cache-Control", "no-cache")
	h.Set("Origin", "https://chat.zalo.me")
	h.Set("Pragma", "no-cache")
	h.Set("User-Agent", util.AsString(firstNonNil(s.State.Headers["User-Agent"], util.HEADERS["User-Agent"])))
	h.Set("Cookie", rawCookies)
	return h, nil
}

func (s *SocketAPI) publishError(err error) {
	if err == nil || s.Hub == nil {
		return
	}
	s.Hub.PublishError(worker.SocketErrorEvent{Err: err, Timestamp: util.Now()})
}

func (s *SocketAPI) dispatch(fn func()) {
	if !s.thread {
		fn()
		return
	}
	if s.dispatchSem == nil {
		s.dispatchSem = make(chan struct{}, 8)
	}
	s.dispatchSem <- struct{}{}
	go func() {
		defer func() { <-s.dispatchSem }()
		fn()
	}()
}

func firstNonNil(values ...any) any {
	for _, v := range values {
		if v == nil {
			continue
		}
		if s, ok := v.(string); ok && s == "" {
			continue
		}
		return v
	}
	return nil
}

func buildPingPayload() []byte {
	payload := map[string]any{"eventId": util.Now()}
	data, _ := json.Marshal(payload)
	buf := bytes.NewBuffer(nil)
	_ = buf.WriteByte(1)
	_ = binary.Write(buf, binary.LittleEndian, uint32(2))
	_ = buf.WriteByte(1)
	_, _ = buf.Write(data)
	return buf.Bytes()
}

func parseJSONBytes(data []byte) map[string]any {
	out, err := util.DecodeJSONMap(data)
	if err != nil {
		return map[string]any{}
	}
	return out
}

func decodeStringJSON(v any) any {
	if s, ok := v.(string); ok {
		if out, err := util.DecodeJSONAny([]byte(s)); err == nil {
			return out
		}
	}
	return v
}

func cookieMap(jar http.CookieJar, rawURL string) map[string]string {
	out := map[string]string{}
	u, err := url.Parse(rawURL)
	if err != nil {
		return out
	}
	for _, c := range jar.Cookies(u) {
		out[c.Name] = c.Value
	}
	return out
}

func qrCookieMap(jar http.CookieJar) map[string]string {
	out := map[string]string{}
	if jar == nil {
		return out
	}
	urls := []string{
		"https://id.zalo.me/account",
		"https://chat.zalo.me/",
		"https://wpa.chat.zalo.me/",
		"https://jr.chat.zalo.me/",
		"https://tt-group-wpa.chat.zalo.me/",
	}
	for _, rawURL := range urls {
		for k, v := range cookieMap(jar, rawURL) {
			if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
				continue
			}
			out[k] = v
		}
	}
	return out
}

func freshClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar, Timeout: 30 * time.Second}
}

var _ core.ThreadType
var _ = base64.StdEncoding
var _ = filepath.Join
var _ = os.MkdirAll

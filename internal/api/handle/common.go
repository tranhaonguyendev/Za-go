package handle

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	core "github.com/tranhaonguyendev/za-go/internal/core"
	"github.com/tranhaonguyendev/za-go/internal/util"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

func (s *SendAPI) normThreadID(threadID string) string {
	threadID = strings.TrimSpace(threadID)
	if threadID == "" {
		return s.UID
	}
	if n, err := strconv.ParseInt(threadID, 10, 64); err == nil {
		return strconv.FormatInt(n, 10)
	}
	return threadID
}

func normalizeClientID(value any) any {
	text := strings.TrimSpace(util.AsString(value))
	if text == "" {
		return nil
	}
	if num, err := strconv.ParseInt(text, 10, 64); err == nil {
		return num
	}
	return text
}

func enrichThreadResponse(decoded any, clientID any) any {
	m, ok := decoded.(map[string]any)
	if !ok {
		return decoded
	}
	normalizedClientID := normalizeClientID(clientID)
	if normalizedClientID == nil {
		return m
	}
	if util.AsString(m["clientId"]) == "" {
		m["clientId"] = normalizedClientID
	}
	if util.AsString(m["cliMsgId"]) == "" {
		m["cliMsgId"] = normalizedClientID
	}
	return m
}

func (s *SendAPI) parseThreadResponse(data map[string]any, threadType core.ThreadType, clientID any) (any, error) {
	decoded, err := util.ParseResponseEnvelope(data, s.State.GetSecretkey())
	if err != nil {
		return nil, err
	}
	decoded = enrichThreadResponse(decoded, clientID)

	switch threadType {
	case core.GROUP:
		return worker.GroupFromDict(decoded), nil
	case core.USER:
		return worker.UserFromDict(decoded), nil
	default:
		return nil, fmt.Errorf("thread type is invalid")
	}
}

func (s *SendAPI) remoteHeadSize(rawURL string) int64 {
	req, err := http.NewRequest(http.MethodHead, rawURL, nil)
	if err != nil {
		return 0
	}
	if ua := s.State.Headers["User-Agent"]; ua != "" {
		req.Header.Set("User-Agent", ua)
	}
	resp, err := s.State.Session.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	return resp.ContentLength
}

func (s *SendAPI) remoteGetBytes(rawURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	if ua := s.State.Headers["User-Agent"]; ua != "" {
		req.Header.Set("User-Agent", ua)
	}
	resp, err := s.State.Session.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func md5Hex(content []byte) string {
	sum := md5.Sum(content)
	return hex.EncodeToString(sum[:])
}

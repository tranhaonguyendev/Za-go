package handle

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/util"
)

func (s *SendAPI) SendVoice(voiceURL string, threadID string, threadType core.ThreadType, fileSize int, ttl int) (any, error) {
	if fileSize == 0 {
		if content, err := s.remoteGetBytes(voiceURL); err == nil {
			fileSize = len(content)
		}
	}

	params := url.Values{}
	params.Set("zpw_ver", "645")
	params.Set("zpw_type", strconv.Itoa(s.APILoginType))
	params.Set("nretry", "0")

	payload := map[string]any{
		"ttl":      ttl,
		"zsource":  -1,
		"msgType":  3,
		"clientId": strconv.FormatInt(util.Now(), 10),
		"msgInfo": util.JSONString(map[string]any{
			"voiceUrl": voiceURL,
			"m4aUrl":   voiceURL,
			"fileSize": fileSize,
		}),
	}

	var endpoint string
	switch threadType {
	case core.USER:
		endpoint = "https://tt-files-wpa.chat.zalo.me/api/message/forward"
		payload["toId"] = threadID
		payload["imei"] = s.IMEI
	case core.GROUP:
		endpoint = "https://tt-files-wpa.chat.zalo.me/api/group/forward"
		payload["visibility"] = 0
		payload["grid"] = threadID
	default:
		return nil, fmt.Errorf("thread type is invalid")
	}

	enc, err := util.ZaloEncode(payload, s.State.GetSecretkey())
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("params", enc)

	resp, err := s.State.PostSession(endpoint+"?"+params.Encode(), form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return s.parseThreadResponse(out, threadType, payload["clientId"])
}

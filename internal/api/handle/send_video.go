package handle

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

func (s *SendAPI) SendVideo(videoURL string, thumbnailURL string, duration int, threadID string, threadType core.ThreadType, width int, height int, message *worker.Message, ttl int) (any, error) {
	if width == 0 {
		width = 1280
	}
	if height == 0 {
		height = 720
	}
	fileSize := int(s.remoteHeadSize(videoURL))

	params := url.Values{}
	params.Set("zpw_ver", "645")
	params.Set("zpw_type", strconv.Itoa(s.APILoginType))
	params.Set("nretry", "0")

	title := ""
	if message != nil {
		title = message.Text
	}

	msgInfo := map[string]any{
		"videoUrl": videoURL,
		"thumbUrl": thumbnailURL,
		"duration": duration,
		"width":    width,
		"height":   height,
		"fileSize": fileSize,
		"properties": map[string]any{
			"color":   -1,
			"size":    -1,
			"type":    1003,
			"subType": 0,
			"ext": map[string]any{
				"sSrcType":         -1,
				"sSrcStr":          "",
				"msg_warning_type": 0,
			},
		},
		"title": title,
	}

	payload := map[string]any{
		"clientId": strconv.FormatInt(util.Now(), 10),
		"ttl":      ttl,
		"zsource":  704,
		"msgType":  5,
		"msgInfo":  util.JSONString(msgInfo),
	}
	if message != nil && message.Mention != "" {
		payload["mentionInfo"] = message.Mention
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

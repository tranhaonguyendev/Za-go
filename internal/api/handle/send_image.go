package handle

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	core "github.com/tranhaonguyendev/za-go/internal/core"
	"github.com/tranhaonguyendev/za-go/internal/util"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

func (s *SendAPI) SendImage(imageURL string, threadID string, threadType core.ThreadType, width int, height int, message *worker.Message, ttl int) (any, error) {
	if width == 0 {
		width = 2560
	}
	if height == 0 {
		height = 2560
	}
	clientID := util.Now()
	desc := ""
	if message != nil {
		desc = message.Text
	}

	params := url.Values{}
	params.Set("zpw_ver", "645")
	params.Set("zpw_type", strconv.Itoa(s.APILoginType))
	params.Set("nretry", "0")

	payload := map[string]any{
		"photoId":   int(clientID * 2),
		"clientId":  int(clientID),
		"desc":      desc,
		"width":     width,
		"height":    height,
		"rawUrl":    imageURL,
		"thumbUrl":  imageURL,
		"hdUrl":     imageURL,
		"thumbSize": "0",
		"fileSize":  "0",
		"hdSize":    "0",
		"zsource":   -1,
		"jcp":       util.JSONString(map[string]any{"sendSource": 1, "convertible": "jxl"}),
		"ttl":       ttl,
		"imei":      s.IMEI,
	}
	if message != nil && message.Mention != "" {
		payload["mentionInfo"] = message.Mention
	}

	var endpoint string
	switch threadType {
	case core.USER:
		endpoint = "https://tt-files-wpa.chat.zalo.me/api/message/photo_original/send"
		payload["toid"] = threadID
		payload["normalUrl"] = imageURL
	case core.GROUP:
		endpoint = "https://tt-files-wpa.chat.zalo.me/api/group/photo_original/send"
		payload["grid"] = threadID
		payload["oriUrl"] = imageURL
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
	return s.parseThreadResponse(out, threadType, clientID)
}

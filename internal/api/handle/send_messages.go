package handle

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	base "github.com/nguyendev/zago/internal/api/common"
	upload "github.com/nguyendev/zago/internal/api/properties/host"
	"github.com/nguyendev/zago/internal/app"
	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

type SendAPI struct {
	*base.BaseAPI
	Uploader *upload.UploadAPI
}

func NewSendAPI(state *app.State, loginType int, hub *worker.Hub) *SendAPI {
	return &SendAPI{BaseAPI: base.NewBaseAPI(state, loginType, hub), Uploader: upload.NewUploadAPI(state, loginType, hub)}
}

func (s *SendAPI) SendMessage(message worker.Message, threadID string, threadType core.ThreadType) (any, error) {
	threadID = s.normThreadID(threadID)
	if strings.TrimSpace(message.ParseMode) != "" {
		parsedMessage, err := worker.NewParsedMessage(message.Text, message.Style, message.Mention, message.ParseMode)
		if err != nil {
			return nil, err
		}
		message = parsedMessage
	}
	params := url.Values{}
	params.Set("zpw_ver", "645")
	params.Set("zpw_type", strconv.Itoa(s.APILoginType))
	params.Set("nretry", "0")

	clientID := util.Now()
	payload := map[string]any{
		"message":  message.Text,
		"clientId": clientID,
		"imei":     s.IMEI,
		"ttl":      0,
	}
	if strings.TrimSpace(message.Style) != "" {
		payload["textProperties"] = message.Style
	}

	var endpoint string
	switch threadType {
	case core.USER:
		endpoint = "https://tt-chat2-wpa.chat.zalo.me/api/message/sms"
		payload["toid"] = threadID
	case core.GROUP:
		payload["grid"] = threadID
		payload["visibility"] = 0
		if strings.TrimSpace(message.Mention) != "" {
			endpoint = "https://tt-group-wpa.chat.zalo.me/api/group/mention"
			payload["mentionInfo"] = message.Mention
		} else {
			endpoint = "https://tt-group-wpa.chat.zalo.me/api/group/sendmsg"
		}
	default:
		return nil, fmt.Errorf("thread type is invalid")
	}
	if threadType == core.USER && strings.TrimSpace(message.Mention) != "" {
		payload["mentionInfo"] = message.Mention
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

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return s.parseThreadResponse(data, threadType, clientID)
}

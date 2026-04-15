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

func (s *SendAPI) SendReaction(messageObject *worker.MessageObject, reactionIcon string, threadID string, threadType core.ThreadType, reactionType int) (any, error) {
	if reactionType == 0 {
		reactionType = 75
	}
	params := url.Values{}
	params.Set("zpw_ver", "647")
	params.Set("zpw_type", strconv.Itoa(s.APILoginType))

	msg := map[string]any{
		"rMsg": []any{map[string]any{
			"gMsgID":  util.AsInt(messageObject.Get("msgId")),
			"cMsgID":  util.AsInt(messageObject.Get("cliMsgId")),
			"msgType": util.GetClientMessageType(util.AsString(messageObject.Get("msgType"))),
		}},
		"rIcon":  reactionIcon,
		"rType":  reactionType,
		"source": 6,
	}

	payload := map[string]any{
		"react_list": []any{map[string]any{
			"message":  util.JSONString(msg),
			"clientId": util.Now(),
		}},
		"imei": s.IMEI,
	}

	var endpoint string
	switch threadType {
	case core.USER:
		endpoint = "https://reaction.chat.zalo.me/api/message/reaction"
		payload["toid"] = threadID
	case core.GROUP:
		endpoint = "https://reaction.chat.zalo.me/api/group/reaction"
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
	return s.parseThreadResponse(out, threadType, util.AsMap(util.AsSlice(payload["react_list"])[0])["clientId"])
}

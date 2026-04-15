package handle

import (
	"fmt"

	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

func (s *SendAPI) SendMultiReaction(messageObject *worker.MessageObject, reactionIcon string, threadID string, threadType core.ThreadType, reactionType int, numreact int) (any, error) {
	if numreact <= 0 {
		numreact = 1
	}
	rMsgItem := map[string]any{"gMsgID": util.AsInt(messageObject.Get("msgId")), "cMsgID": util.AsInt(messageObject.Get("cliMsgId")), "msgType": util.GetClientMessageType(util.AsString(messageObject.Get("msgType")))}
	rMsgs := make([]any, 0, numreact)
	for i := 0; i < numreact; i++ {
		rMsgs = append(rMsgs, rMsgItem)
	}
	msg := map[string]any{"rMsg": rMsgs, "rIcon": reactionIcon, "rType": reactionType, "source": 6}
	payload := map[string]any{"react_list": []any{map[string]any{"message": util.JSONString(msg), "clientId": util.Now()}}, "imei": s.IMEIValue()}
	url := "https://reaction.chat.zalo.me/api/message/reaction"
	switch threadType {
	case core.USER:
		payload["toid"] = threadID
	case core.GROUP:
		url = "https://reaction.chat.zalo.me/api/group/reaction"
		payload["grid"] = threadID
	default:
		return nil, fmt.Errorf("thread type is invalid")
	}
	form, err := s.EncodedForm(payload)
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON(url, s.Query(map[string]any{"zpw_ver": 647, "zpw_type": s.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return s.ParseThread(data, threadType)
}

package properties

import (
	"fmt"

	core "github.com/tranhaonguyendev/za-go/internal/core"
	"github.com/tranhaonguyendev/za-go/internal/util"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

func (p *PropertiesAPI) MarkAsRead(msgID any, cliMsgID any, senderID any, threadID string, threadType core.ThreadType, method string) (bool, error) {
	if method == "" {
		method = "webchat"
	}
	dest := "0"
	if threadType == core.GROUP {
		dest = threadID
	}
	entry := map[string]any{"cmi": util.AsString(cliMsgID), "gmi": util.AsString(msgID), "si": util.AsString(senderID), "di": dest, "mt": method, "st": 3, "ts": util.AsString(util.Now())}
	info := map[string]any{"data": []any{entry}}
	payload := map[string]any{"msgInfos": util.JSONString(info), "imei": p.IMEIValue()}
	url := "https://tt-chat1-wpa.chat.zalo.me/api/message/seenv2"
	if threadType == core.USER {
		entry["at"] = 7
		entry["cmd"] = 501
		payload["senderId"] = dest
	} else if threadType == core.GROUP {
		url = "https://tt-group-wpa.chat.zalo.me/api/group/seenv2"
		entry["at"] = 0
		entry["cmd"] = 511
		payload["grid"] = dest
	} else {
		return false, fmt.Errorf("thread type is invalid")
	}
	form, err := p.EncodedForm(payload)
	if err != nil {
		return false, err
	}
	data, err := p.PostJSON(url, p.Query(map[string]any{"zpw_ver": 645, "zpw_type": p.APILoginType, "nretry": 0}), form)
	if err != nil {
		return false, err
	}
	if util.AsInt(data["error_code"]) != 0 {
		return false, fmt.Errorf("error #%d when sending requests: %s", util.AsInt(data["error_code"]), util.AsString(data["error_message"]))
	}
	if p.Hub != nil {
		p.Hub.PublishSeen(worker.SeenEvent{MsgIDs: msgID, ThreadID: threadID, ThreadType: threadType, Timestamp: util.Now()})
	}
	return true, nil
}

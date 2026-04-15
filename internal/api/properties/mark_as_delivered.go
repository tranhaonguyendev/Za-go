package properties

import (
	"fmt"

	core "github.com/tranhaonguyendev/za-go/internal/core"
	"github.com/tranhaonguyendev/za-go/internal/util"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

func (p *PropertiesAPI) MarkAsDelivered(msgID any, cliMsgID any, senderID any, threadID string, threadType core.ThreadType, method string) (bool, error) {
	if method == "" {
		method = "webchat"
	}
	dest := "0"
	if threadType == core.GROUP {
		dest = threadID
	}
	info := map[string]any{
		"seen": 0,
		"data": []any{map[string]any{"cmi": util.AsString(cliMsgID), "gmi": util.AsString(msgID), "si": util.AsString(senderID), "di": dest, "mt": method, "st": 3, "at": 0, "ts": util.AsString(util.Now())}},
	}
	payload := map[string]any{"msgInfos": util.JSONString(info)}
	url := "https://tt-chat3-wpa.chat.zalo.me/api/message/deliveredv2"
	if threadType == core.USER {
		util.AsSlice(info["data"])[0].(map[string]any)["cmd"] = 501
	} else {
		url = "https://tt-group-wpa.chat.zalo.me/api/group/deliveredv2"
		util.AsSlice(info["data"])[0].(map[string]any)["cmd"] = 521
		info["grid"] = dest
		payload["imei"] = p.IMEIValue()
	}
	form, err := p.EncodedForm(payload)
	if err != nil {
		return false, err
	}
	data, err := p.PostJSON(url, p.Query(map[string]any{"zpw_ver": 645, "zpw_type": p.APILoginType}), form)
	if err != nil {
		return false, err
	}
	if util.AsInt(data["error_code"]) != 0 {
		return false, fmt.Errorf("error #%d when sending requests: %s", util.AsInt(data["error_code"]), util.AsString(data["error_message"]))
	}
	if p.Hub != nil {
		p.Hub.PublishDelivery(worker.DeliveryEvent{MsgIDs: msgID, ThreadID: threadID, ThreadType: threadType, Timestamp: util.Now()})
	}
	return true, nil
}

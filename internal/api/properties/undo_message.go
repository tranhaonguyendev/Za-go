package properties

import (
	"fmt"

	core "github.com/tranhaonguyendev/za-go/internal/core"
	"github.com/tranhaonguyendev/za-go/internal/util"
)

func (p *PropertiesAPI) UndoMessage(msgID any, cliMsgID any, threadID string, threadType core.ThreadType) (any, error) {
	payload := map[string]any{"msgId": fmt.Sprintf("%v", msgID), "cliMsgIdUndo": fmt.Sprintf("%v", cliMsgID), "clientId": util.Now()}
	url := "https://tt-chat3-wpa.chat.zalo.me/api/message/undo"
	switch threadType {
	case core.USER:
		payload["toid"] = threadID
	case core.GROUP:
		url = "https://tt-group-wpa.chat.zalo.me/api/group/undomsg"
		payload["grid"] = threadID
		payload["visibility"] = 0
		payload["imei"] = p.IMEIValue()
	default:
		return nil, fmt.Errorf("thread type is invalid")
	}
	form, err := p.EncodedForm(payload)
	if err != nil {
		return nil, err
	}
	data, err := p.PostJSON(url, p.Query(map[string]any{"zpw_ver": 645, "zpw_type": p.APILoginType, "nretry": 0}), form)
	if err != nil {
		return nil, err
	}
	return p.ParseThread(data, threadType)
}

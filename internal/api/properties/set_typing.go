package properties

import (
	"fmt"
	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/util"
)

func (p *PropertiesAPI) SetTyping(threadID string, threadType core.ThreadType) (bool, error) {
	payload := map[string]any{"imei": p.IMEIValue()}
	url := "https://tt-chat1-wpa.chat.zalo.me/api/message/typing"
	if threadType == core.USER {
		payload["toid"] = threadID
		payload["destType"] = 3
	} else if threadType == core.GROUP {
		url = "https://tt-group-wpa.chat.zalo.me/api/group/typing"
		payload["grid"] = threadID
	} else {
		return false, fmt.Errorf("thread type is invalid")
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
	return true, nil
}

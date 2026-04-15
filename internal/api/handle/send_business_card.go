package handle

import (
	"fmt"

	core "github.com/tranhaonguyendev/za-go/internal/core"
	"github.com/tranhaonguyendev/za-go/internal/util"
)

func (s *SendAPI) SendBusinessCard(userID string, qrCodeURL string, threadID string, threadType core.ThreadType, phone string, ttl int) (any, error) {
	msgInfo := map[string]any{"contactUid": userID, "qrCodeUrl": qrCodeURL}
	if phone != "" {
		msgInfo["phone"] = phone
	}
	payload := map[string]any{"ttl": ttl, "msgType": 6, "clientId": util.AsString(util.Now()), "msgInfo": util.JSONString(msgInfo)}
	url := "https://tt-files-wpa.chat.zalo.me/api/message/forward"
	switch threadType {
	case core.USER:
		payload["toId"] = threadID
		payload["imei"] = s.IMEIValue()
	case core.GROUP:
		url = "https://tt-files-wpa.chat.zalo.me/api/group/forward"
		payload["visibility"] = 0
		payload["grid"] = threadID
	default:
		return nil, fmt.Errorf("thread type is invalid")
	}
	form, err := s.EncodedForm(payload)
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON(url, s.Query(map[string]any{"zpw_ver": 645, "zpw_type": s.APILoginType, "nretry": 0}), form)
	if err != nil {
		return nil, err
	}
	return s.ParseThread(data, threadType)
}

package handle

import (
	"fmt"

	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/util"
)

func (s *SendAPI) SendSticker(stickerType int, stickerID int, cateID int, threadID string, threadType core.ThreadType, ttl int) (any, error) {
	payload := map[string]any{"stickerId": stickerID, "cateId": cateID, "type": stickerType, "clientId": util.Now(), "imei": s.IMEIValue(), "ttl": ttl}
	url := "https://tt-chat2-wpa.chat.zalo.me/api/message/sticker"
	switch threadType {
	case core.USER:
		payload["zsource"] = 106
		payload["toid"] = threadID
	case core.GROUP:
		url = "https://tt-group-wpa.chat.zalo.me/api/group/sticker"
		payload["zsource"] = 103
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

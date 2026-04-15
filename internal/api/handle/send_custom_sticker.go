package handle

import (
	"fmt"

	core "github.com/tranhaonguyendev/za-go/internal/core"
	"github.com/tranhaonguyendev/za-go/internal/util"
)

func (s *SendAPI) SendCustomSticker(staticImgURL string, animationImgURL string, threadID string, threadType core.ThreadType, reply string, width int, height int, ttl int, ai bool) (any, error) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	payload := map[string]any{
		"clientId":     util.Now(),
		"title":        "",
		"oriUrl":       staticImgURL,
		"thumbUrl":     staticImgURL,
		"hdUrl":        staticImgURL,
		"width":        width,
		"height":       height,
		"properties":   util.JSONString(map[string]any{"subType": 0, "color": -1, "size": -1, "type": 3, "ext": util.JSONString(map[string]any{"sSrcStr": "", "sSrcType": -1})}),
		"contentId":    util.Now(),
		"thumb_height": width,
		"thumb_width":  height,
		"webp":         util.JSONString(map[string]any{"width": width, "height": height, "url": animationImgURL}),
		"zsource":      -1,
		"ttl":          ttl,
	}
	if ai {
		payload["jcp"] = util.JSONString(map[string]any{"pStickerType": 1})
	}
	if reply != "" {
		payload["refMessage"] = reply
	}
	url := "https://tt-files-wpa.chat.zalo.me/api/message/photo_url"
	switch threadType {
	case core.USER:
		payload["toId"] = threadID
	case core.GROUP:
		url = "https://tt-files-wpa.chat.zalo.me/api/group/photo_url"
		payload["visibility"] = 0
		payload["grid"] = threadID
	default:
		return nil, fmt.Errorf("thread type is invalid")
	}
	form, err := s.EncodedForm(payload)
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON(url, s.Query(map[string]any{"zpw_ver": 669, "zpw_type": s.APILoginType, "nretry": 0}), form)
	if err != nil {
		return nil, err
	}
	return s.ParseThread(data, threadType)
}

package get

import (
	"net/url"
	"strconv"

	"github.com/nguyendev/zago/internal/util"
)

func (g *GetAPI) UpdateStickersPersonal(cateIDs any, version int) (any, error) {
	ids := []int{}
	switch t := cateIDs.(type) {
	case []int:
		ids = append(ids, t...)
	case []string:
		for _, item := range t {
			if n, err := strconv.Atoi(item); err == nil {
				ids = append(ids, n)
			}
		}
	case []any:
		for _, item := range t {
			ids = append(ids, util.AsInt(item))
		}
	default:
		if cateIDs != nil {
			ids = append(ids, util.AsInt(cateIDs))
		}
	}
	enc, err := g.Encode(map[string]any{"version": version, "sticker_cates": ids, "imei": g.IMEIValue()})
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("zpw_ver", "645")
	q.Set("zpw_type", strconv.Itoa(g.APILoginType))
	q.Set("params", enc)
	data, err := g.GetJSON("https://tt-sticker-wpa.chat.zalo.me/api/message/sticker/personalized/update", q)
	if err != nil {
		return nil, err
	}
	return g.ParseRaw(data)
}

func (g *GetAPI) UpdatePersonalSticker(cateIDs any, version int) (any, error) {
	return g.UpdateStickersPersonal(cateIDs, version)
}

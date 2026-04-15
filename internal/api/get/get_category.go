package get

import (
	"net/url"
	"strconv"
)

func (g *GetAPI) GetCategory(catesID any) (any, error) {
	enc, err := g.Encode(map[string]any{"cid": catesID})
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("zpw_ver", "645")
	q.Set("zpw_type", strconv.Itoa(g.APILoginType))
	q.Set("params", enc)
	data, err := g.GetJSON("https://tt-sticker-wpa.chat.zalo.me/api/message/sticker/category/sticker_detail", q)
	if err != nil {
		return nil, err
	}
	return g.ParseRaw(data)
}

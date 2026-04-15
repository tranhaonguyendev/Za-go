package get

import (
	"net/url"
	"strconv"
)

func (g *GetAPI) FindSticker(keyword string) (any, error) {
	enc, err := g.Encode(map[string]any{"keyword": keyword, "gif": 1, "guggy": 0, "imei": g.IMEIValue()})
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("zpw_ver", "678")
	q.Set("zpw_type", strconv.Itoa(g.APILoginType))
	q.Set("params", enc)
	data, err := g.GetJSON("https://tt-sticker-wpa.chat.zalo.me/api/message/sticker/suggest/stickers", q)
	if err != nil {
		return nil, err
	}
	return g.ParseRaw(data)
}

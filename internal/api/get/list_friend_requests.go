package get

import (
	"net/url"
	"strconv"

	"github.com/nguyendev/zago/internal/util"
)

func (g *GetAPI) ListFriendRequests() (any, error) {
	enc, err := util.ZaloEncode(map[string]any{"imei": g.IMEI}, g.State.GetSecretkey())
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("zpw_ver", "664")
	q.Set("zpw_type", strconv.Itoa(g.APILoginType))

	form := url.Values{}
	form.Set("params", enc)

	data, err := g.postJSON("https://tt-friend-wpa.chat.zalo.me/api/friend/recommendsv2/list?"+q.Encode(), form)
	if err != nil {
		return nil, err
	}
	return g.decodeAPIData(data)
}

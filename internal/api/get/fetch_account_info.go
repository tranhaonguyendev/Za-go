package get

import (
	"net/url"
	"strconv"

	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

func (g *GetAPI) FetchAccountInfo() (any, error) {
	enc, err := util.ZaloEncode(map[string]any{
		"avatar_size": 120,
		"imei":        g.IMEI,
	}, g.State.GetSecretkey())
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("params", enc)
	q.Set("zpw_ver", "645")
	q.Set("zpw_type", strconv.Itoa(g.APILoginType))
	q.Set("os", "8")
	q.Set("browser", "0")

	data, err := g.getJSON("https://tt-profile-wpa.chat.zalo.me/api/social/profile/me-v2?" + q.Encode())
	if err != nil {
		return nil, err
	}
	decoded, err := g.decodeAPIData(data)
	if err != nil {
		return nil, err
	}
	return worker.UserFromDict(decoded), nil
}

package get

import (
	"net/url"
	"strconv"

	"github.com/tranhaonguyendev/za-go/internal/util"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

func (g *GetAPI) GetRecentGroup(groupID string) (any, error) {
	enc, err := util.ZaloEncode(map[string]any{
		"groupId":     groupID,
		"globalMsgId": float64(10000000000000000),
		"count":       50,
		"msgIds":      []any{},
		"imei":        g.IMEI,
		"src":         1,
	}, g.State.GetSecretkey())
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("params", enc)
	q.Set("zpw_ver", "645")
	q.Set("zpw_type", strconv.Itoa(g.APILoginType))
	q.Set("nretry", "0")

	data, err := g.getJSON("https://tt-group-cm.chat.zalo.me/api/cm/getrecentv2?" + q.Encode())
	if err != nil {
		return nil, err
	}
	decoded, err := g.decodeAPIData(data)
	if err != nil {
		return nil, err
	}
	return worker.GroupFromDict(decoded), nil
}

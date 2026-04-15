package get

import (
	"net/url"
	"strconv"

	"github.com/tranhaonguyendev/za-go/internal/worker"
)

func (g *GetAPI) FetchAllGroups() (any, error) {
	q := url.Values{}
	q.Set("zpw_ver", "645")
	q.Set("zpw_type", strconv.Itoa(g.APILoginType))

	data, err := g.getJSON("https://tt-group-wpa.chat.zalo.me/api/group/getlg/v4?" + q.Encode())
	if err != nil {
		return nil, err
	}
	decoded, err := g.decodeAPIData(data)
	if err != nil {
		return nil, err
	}
	return worker.GroupFromDict(decoded), nil
}

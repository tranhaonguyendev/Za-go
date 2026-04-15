package get

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

func (g *GetAPI) FetchGroupInfo(groupIDs ...string) (any, error) {
	gridMap := map[string]int{}
	for _, id := range groupIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			gridMap[id] = 0
		}
	}
	gridJSON, _ := json.Marshal(gridMap)

	enc, err := util.ZaloEncode(map[string]any{
		"gridVerMap": string(gridJSON),
	}, g.State.GetSecretkey())
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("zpw_ver", "645")
	q.Set("zpw_type", strconv.Itoa(g.APILoginType))

	form := url.Values{}
	form.Set("params", enc)

	data, err := g.postJSON("https://tt-group-wpa.chat.zalo.me/api/group/getmg-v2?"+q.Encode(), form)
	if err != nil {
		return nil, err
	}
	decoded, err := g.decodeAPIData(data)
	if err != nil {
		return nil, err
	}
	return worker.GroupFromDict(decoded), nil
}

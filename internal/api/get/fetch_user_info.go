package get

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

func (g *GetAPI) FetchUserInfo(userIDs ...string) (any, error) {
	friendMap := make([]string, 0, len(userIDs))
	for _, id := range userIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			friendMap = append(friendMap, id+"_0")
		}
	}

	enc, err := util.ZaloEncode(map[string]any{
		"phonebook_version":   util.Now() / 1000,
		"friend_pversion_map": friendMap,
		"avatar_size":         120,
		"language":            "vi",
		"show_online_status":  1,
		"imei":                g.IMEI,
	}, g.State.GetSecretkey())
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("zpw_ver", "645")
	q.Set("zpw_type", strconv.Itoa(g.APILoginType))

	form := url.Values{}
	form.Set("params", enc)

	data, err := g.postJSON("https://tt-profile-wpa.chat.zalo.me/api/social/friend/getprofiles/v2?"+q.Encode(), form)
	if err != nil {
		return nil, err
	}
	decoded, err := g.decodeAPIData(data)
	if err != nil {
		return nil, err
	}
	return worker.UserFromDict(decoded), nil
}

package get

import (
	"net/url"
	"strconv"

	"github.com/tranhaonguyendev/za-go/internal/util"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

func (g *GetAPI) FetchAllFriends() (any, error) {
	enc, err := util.ZaloEncode(map[string]any{
		"incInvalid":  0,
		"page":        1,
		"count":       20000,
		"avatar_size": 120,
		"actiontime":  0,
	}, g.State.GetSecretkey())
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("params", enc)
	q.Set("zpw_ver", "645")
	q.Set("zpw_type", strconv.Itoa(g.APILoginType))
	q.Set("nretry", "0")

	data, err := g.getJSON("https://profile-wpa.chat.zalo.me/api/social/friend/getfriends?" + q.Encode())
	if err != nil {
		return nil, err
	}
	decoded, err := g.decodeAPIData(data)
	if err != nil {
		return nil, err
	}

	items, ok := decoded.([]any)
	if !ok {
		return []*worker.User{}, nil
	}
	users := make([]*worker.User, 0, len(items))
	for _, item := range items {
		users = append(users, worker.UserFromDict(item))
	}
	return users, nil
}

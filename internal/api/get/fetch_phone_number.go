package get

import (
	"net/url"
	"strconv"

	"github.com/tranhaonguyendev/za-go/internal/util"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

func (g *GetAPI) FetchPhoneNumber(phoneNumber string, language string) (*worker.User, error) {
	if language == "" {
		language = "vi"
	}
	payload := map[string]any{
		"phone":       util.NormalizePhone(phoneNumber),
		"avatar_size": 240,
		"language":    language,
		"imei":        g.IMEIValue(),
		"reqSrc":      85,
	}
	enc, err := g.Encode(payload)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("zpw_ver", "645")
	q.Set("zpw_type", strconv.Itoa(g.APILoginType))
	q.Set("params", enc)
	data, err := g.GetJSON("https://tt-friend-wpa.chat.zalo.me/api/friend/profile/get", q)
	if err != nil {
		return nil, err
	}
	decoded, err := g.ParseRaw(data)
	if err != nil {
		return nil, err
	}
	return worker.UserFromDict(decoded), nil
}

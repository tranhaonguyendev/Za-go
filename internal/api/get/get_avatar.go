package get

import (
	"fmt"

	"github.com/tranhaonguyendev/za-go/internal/util"
)

func (g *GetAPI) GetAvatar(userID string, avatarSize int) (any, error) {
	if avatarSize == 0 {
		avatarSize = 120
	}
	form, err := g.EncodedForm(map[string]any{"fid": userID, "imei": g.IMEIValue(), "avatar_size": avatarSize})
	if err != nil {
		return nil, err
	}
	q := g.Query(map[string]any{"zpw_ver": 645, "zpw_type": g.APILoginType})
	data, err := g.PostJSON("https://tt-profile-wpa.chat.zalo.me/api/social/profile/avatar", q, form)
	if err != nil {
		return nil, err
	}
	if util.AsInt(data["error_code"]) != 0 {
		return nil, fmt.Errorf("error #%d when sending requests: %s", util.AsInt(data["error_code"]), util.AsString(data["error_message"]))
	}
	return g.NormalizeMaybeJSON(data["data"]), nil
}

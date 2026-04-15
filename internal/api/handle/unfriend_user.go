package handle

import (
	"fmt"

	"github.com/nguyendev/zago/internal/util"
)

func (s *SendAPI) UnfriendUser(userID string, language string) (map[string]any, error) {
	if language == "" {
		language = "en"
	}
	form, err := s.EncodedForm(map[string]any{"fid": userID, "language": language})
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON("https://tt-friend-wpa.chat.zalo.me/api/friend/remove", s.Query(map[string]any{"zpw_ver": 641, "zpw_type": s.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	if util.AsInt(data["error_code"]) == 0 {
		return map[string]any{"status": "success", "message": "Unfriended successfully."}, nil
	}
	return nil, fmt.Errorf("error #%d when unfriending: %s", util.AsInt(data["error_code"]), util.AsString(data["error_message"]))
}

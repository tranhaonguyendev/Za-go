package handle

import (
	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

func (s *SendAPI) AddFriend(userID string, msg string, language string) (*worker.User, error) {
	if language == "" {
		language = "en"
	}
	form, err := s.EncodedForm(map[string]any{"toid": userID, "msg": msg, "reqsrc": 30, "imei": s.IMEIValue(), "language": language, "srcParams": util.JSONString(map[string]any{"uidTo": userID})})
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON("https://tt-friend-wpa.chat.zalo.me/api/friend/sendreq", s.Query(map[string]any{"zpw_ver": 645, "zpw_type": s.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return s.ParseUser(data)
}

func (s *SendAPI) SendFriendRequest(userID string, msg string, language string) (*worker.User, error) {
	return s.AddFriend(userID, msg, language)
}

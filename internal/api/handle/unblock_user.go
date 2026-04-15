package handle

import "github.com/nguyendev/zago/internal/worker"

func (s *SendAPI) UnblockUser(userID string) (*worker.User, error) {
	form, err := s.EncodedForm(map[string]any{"fid": userID, "imei": s.IMEIValue()})
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON("https://tt-friend-wpa.chat.zalo.me/api/friend/unblock", s.Query(map[string]any{"zpw_ver": 645, "zpw_type": s.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return s.ParseUser(data)
}

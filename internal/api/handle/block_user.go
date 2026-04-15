package handle

import "github.com/tranhaonguyendev/za-go/internal/worker"

func (s *SendAPI) BlockUser(userID string) (*worker.User, error) {
	form, err := s.EncodedForm(map[string]any{"fid": userID, "imei": s.IMEIValue()})
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON("https://tt-friend-wpa.chat.zalo.me/api/friend/block", s.Query(map[string]any{"zpw_ver": 645, "zpw_type": s.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return s.ParseUser(data)
}

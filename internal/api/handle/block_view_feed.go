package handle

import "github.com/tranhaonguyendev/za-go/internal/worker"

func (s *SendAPI) BlockViewFeed(userID string, isBlockFeed bool) (*worker.User, error) {
	form, err := s.EncodedForm(map[string]any{"fid": userID, "isBlockFeed": map[bool]int{true: 1, false: 0}[isBlockFeed], "imei": s.IMEIValue()})
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON("https://tt-friend-wpa.chat.zalo.me/api/friend/feed/block", s.Query(map[string]any{"zpw_ver": 645, "zpw_type": s.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return s.ParseUser(data)
}

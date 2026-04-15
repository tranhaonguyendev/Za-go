package properties

import (
	"github.com/nguyendev/zago/internal/worker"
)

func (p *PropertiesAPI) AcceptFriendRequest(userID string, language string) (*worker.User, error) {
	if language == "" {
		language = "en"
	}
	form, err := p.EncodedForm(map[string]any{"fid": userID, "language": language})
	if err != nil {
		return nil, err
	}
	data, err := p.PostJSON("https://tt-friend-wpa.chat.zalo.me/api/friend/accept", p.Query(map[string]any{"zpw_ver": 645, "zpw_type": p.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return p.ParseUser(data)
}

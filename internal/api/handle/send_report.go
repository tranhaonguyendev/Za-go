package handle

import (
	"fmt"

	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/util"
)

func (s *SendAPI) SendReport(userID string, threadType core.ThreadType, reason int, content string) (any, error) {
	payload := map[string]any{"idTo": userID, "objId": "person.profile"}
	if content != "" {
		payload["content"] = content
	}
	if content != "" && reason == 0 {
		reason = 0
	}
	if content == "" && reason == 0 {
		reason = 1 + util.AsInt(util.RandomInt()[0])%3
	}
	payload["reason"] = fmt.Sprintf("%d", reason)
	form, err := s.EncodedForm(payload)
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON("https://tt-profile-wpa.chat.zalo.me/api/report/abuse-v2", s.Query(map[string]any{"zpw_ver": 645, "zpw_type": s.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return s.ParseThread(data, threadType)
}

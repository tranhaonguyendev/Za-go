package properties

import (
	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

func (p *PropertiesAPI) ChangeAccountSetting(name string, dob string, gender int, biz map[string]any, language string) (*worker.User, error) {
	if biz == nil {
		biz = map[string]any{}
	}
	if language == "" {
		language = "vi"
	}
	form, err := p.EncodedForm(map[string]any{
		"profile":  util.JSONString(map[string]any{"name": name, "dob": dob, "gender": gender}),
		"biz":      util.JSONString(biz),
		"language": language,
	})
	if err != nil {
		return nil, err
	}
	data, err := p.PostJSON("https://tt-profile-wpa.chat.zalo.me/api/social/profile/update", p.Query(map[string]any{"zpw_ver": 645, "zpw_type": p.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return p.ParseUser(data)
}

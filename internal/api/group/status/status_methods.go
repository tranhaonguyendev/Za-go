package status

import (
	"fmt"
	"time"

	core "github.com/tranhaonguyendev/za-go/internal/core"
	"github.com/tranhaonguyendev/za-go/internal/util"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

func (s *StatusAPI) BoxInviteAccept(groupID string, lang string) (any, error) {
	if lang == "" {
		lang = "en"
	}
	form, err := s.EncodedForm(map[string]any{"grid": util.AsInt64(groupID), "lang": lang})
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON("https://tt-group-wpa.chat.zalo.me/api/group/inv-box/join", s.Query(map[string]any{"zpw_ver": 664, "zpw_type": s.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	if util.AsInt(data["error_code"]) != 0 {
		return nil, fmt.Errorf("error #%d - %s", util.AsInt(data["error_code"]), util.AsString(firstNonNil(data["error_message"], data["data"])))
	}
	return data["data"], nil
}

func (s *StatusAPI) DisableLink(grid string) map[string]any {
	enc, err := s.Encode(map[string]any{"grid": grid})
	if err != nil {
		return map[string]any{"success": false, "error_message": err.Error()}
	}
	data, err := s.PostJSON("https://tt-group-wpa.chat.zalo.me/api/group/link/disable", nil, s.Query(map[string]any{"zpw_ver": 650, "zpw_type": s.APILoginType, "params": enc}))
	if err != nil {
		return map[string]any{"success": false, "error_message": err.Error()}
	}
	if util.AsInt(data["error_code"]) == 0 {
		return map[string]any{"success": true, "message": "Đã vô hiệu hóa liên kết nhóm thành công."}
	}
	return map[string]any{"success": false, "error_code": data["error_code"], "error_message": util.AsString(data["error_message"])}
}

func (s *StatusAPI) DisperseGroup(groupID string) (*worker.Group, error) {
	form, err := s.EncodedForm(map[string]any{"grid": groupID, "imei": s.IMEIValue()})
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON("https://tt-group-wpa.chat.zalo.me/api/group/disperse", s.Query(map[string]any{"zpw_ver": 645, "zpw_type": s.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return s.parseGroup(data)
}

func (s *StatusAPI) GenerateNewLink(grid string) map[string]any {
	enc, err := s.Encode(map[string]any{"grid": grid})
	if err != nil {
		return map[string]any{"success": false, "error_message": err.Error()}
	}
	data, err := s.PostJSON("https://tt-group-wpa.chat.zalo.me/api/group/link/new", nil, s.Query(map[string]any{"zpw_ver": 650, "zpw_type": s.APILoginType, "params": enc}))
	if err != nil {
		return map[string]any{"success": false, "error_message": err.Error()}
	}
	if util.AsInt(data["error_code"]) != 0 {
		return map[string]any{"success": false, "error_code": data["error_code"], "error_message": util.AsString(data["error_message"])}
	}
	decoded, derr := s.ParseRaw(data)
	if derr != nil {
		return map[string]any{"success": false, "error_code": 1337, "error_message": derr.Error()}
	}
	link := util.AsString(util.AsMap(decoded)["link"])
	if link == "" {
		return map[string]any{"success": false, "error_code": 1337, "error_message": ""}
	}
	return map[string]any{"success": true, "new_link": link}
}

func (s *StatusAPI) HandleGroupPending(members any, groupID string, isApprove bool) (*worker.Group, error) {
	ms := []string{}
	switch t := members.(type) {
	case []string:
		ms = append(ms, t...)
	case []any:
		for _, item := range t {
			ms = append(ms, util.AsString(item))
		}
	default:
		ms = append(ms, util.AsString(t))
	}
	enc, err := s.Encode(map[string]any{"grid": groupID, "members": ms, "isApprove": map[bool]int{true: 1, false: 0}[isApprove]})
	if err != nil {
		return nil, err
	}
	data, err := s.GetJSON("https://tt-group-wpa.chat.zalo.me/api/group/pending-mems/review", s.Query(map[string]any{"params": enc, "zpw_ver": 645, "zpw_type": s.APILoginType}))
	if err != nil {
		return nil, err
	}
	return s.parseGroup(data)
}

func (s *StatusAPI) ListInviteBox(page int, invPerPage int, mcount int, lastGroupID any) (any, error) {
	form, err := s.EncodedForm(map[string]any{"mpage": 1, "page": page, "invPerPage": invPerPage, "mcount": mcount, "lastGroupId": lastGroupID, "avatar_size": 120, "member_avatar_size": 120})
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON("https://tt-group-wpa.chat.zalo.me/api/group/inv-box/list", s.Query(map[string]any{"zpw_ver": 664, "zpw_type": s.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	if util.AsInt(data["error_code"]) != 0 {
		return nil, fmt.Errorf("error #%d - %s", util.AsInt(data["error_code"]), util.AsString(firstNonNil(data["error_message"], data["data"])))
	}
	return data["data"], nil
}

func (s *StatusAPI) SetMute(groupID string, mute bool) (any, error) {
	action := 3
	if mute {
		action = 1
	}
	form, err := s.EncodedForm(map[string]any{"toid": groupID, "duration": -1, "action": action, "startTime": time.Now().Unix(), "muteType": 2, "imei": s.IMEIValue()})
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON("https://tt-profile-wpa.chat.zalo.me/api/social/profile/setmute", s.Query(map[string]any{"zpw_ver": 664, "zpw_type": s.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	if util.AsInt(data["error_code"]) != 0 {
		return nil, fmt.Errorf("error #%d - %s", util.AsInt(data["error_code"]), util.AsString(firstNonNil(data["error_message"], data["data"])))
	}
	if raw := data["data"]; raw != nil {
		if _, ok := raw.(map[string]any); ok {
			return raw, nil
		}
		return util.DecodeAPIData(raw, s.Secret())
	}
	return nil, fmt.Errorf("error #1337: data is nil")
}

func (s *StatusAPI) UpdateAutoDeleteChat(ttl int, threadID string, threadType core.ThreadType) (any, error) {
	form, err := s.EncodedForm(map[string]any{"threadId": threadID, "isGroup": map[bool]int{true: 1, false: 0}[threadType == core.GROUP], "ttl": ttl, "clientLang": firstNonNil(s.Language, "vi")})
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON("https://tt-convers-wpa.chat.zalo.me/api/conv/autodelete/updateConvers", s.Query(map[string]any{"zpw_ver": 645, "zpw_type": s.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return s.ParseRaw(data)
}

func (s *StatusAPI) ViewGroupPending(groupID string) (*worker.Group, error) {
	enc, err := s.Encode(map[string]any{"grid": groupID, "imei": s.IMEIValue()})
	if err != nil {
		return nil, err
	}
	data, err := s.GetJSON("https://tt-group-wpa.chat.zalo.me/api/group/pending-mems/list", s.Query(map[string]any{"params": enc, "zpw_ver": 645, "zpw_type": s.APILoginType}))
	if err != nil {
		return nil, err
	}
	return s.parseGroup(data)
}

func (s *StatusAPI) ViewPollDetail(pollID int64) (*worker.Group, error) {
	enc, err := s.Encode(map[string]any{"poll_id": pollID, "imei": s.IMEIValue()})
	if err != nil {
		return nil, err
	}
	data, err := s.GetJSON("https://tt-group-wpa.chat.zalo.me/api/poll/detail", s.Query(map[string]any{"params": enc, "zpw_ver": 645, "zpw_type": s.APILoginType}))
	if err != nil {
		return nil, err
	}
	return s.parseGroup(data)
}

func firstNonNil(values ...any) any {
	for _, v := range values {
		if v == nil {
			continue
		}
		if s, ok := v.(string); ok && s == "" {
			continue
		}
		return v
	}
	return nil
}

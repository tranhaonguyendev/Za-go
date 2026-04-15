package group

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/tranhaonguyendev/za-go/internal/app"
	"github.com/tranhaonguyendev/za-go/internal/util"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

func (g *GroupAPI) AddUsersToGroup(userIDs any, groupID string) (*worker.Group, error) {
	members := g.membersOf(userIDs)
	memberTypes := make([]int, len(members))
	for i := range memberTypes {
		memberTypes[i] = -1
	}
	form, err := g.EncodedForm(map[string]any{"grid": groupID, "members": members, "memberTypes": memberTypes, "imei": g.IMEIValue(), "clientLang": "vi"})
	if err != nil {
		return nil, err
	}
	data, err := g.PostJSON("https://tt-group-wpa.chat.zalo.me/api/group/invite/v2", g.Query(map[string]any{"zpw_ver": 645, "zpw_type": g.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return g.parseGroupResult(data)
}

func (g *GroupAPI) ChangeGroupAvatar(filePath string, groupID string) (*worker.Group, error) {
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	query := g.Query(map[string]any{
		"zpw_ver":  645,
		"zpw_type": g.APILoginType,
		"params": func() string {
			enc, _ := g.Encode(map[string]any{"grid": groupID, "avatarSize": 120, "clientId": "g" + groupID + util.FormatTime("%H:%M %d/%m/%Y"), "originWidth": 640, "originHeight": 640, "imei": g.IMEIValue()})
			return enc
		}(),
	})
	files := []app.MultipartFile{{FieldName: "fileContent", FileName: filepath.Base(filePath), Content: buf, ContentType: "application/octet-stream"}}
	data, err := g.PostMultipartJSON("https://tt-files-wpa.chat.zalo.me/api/group/upavatar", query, nil, files, 15*time.Second)
	if err != nil {
		return nil, err
	}
	return g.parseGroupResult(data)
}

func (g *GroupAPI) ChangeGroupName(groupName string, groupID string) (*worker.Group, error) {
	form, err := g.EncodedForm(map[string]any{"gname": groupName, "grid": groupID})
	if err != nil {
		return nil, err
	}
	data, err := g.PostJSON("https://tt-group-wpa.chat.zalo.me/api/group/updateinfo", g.Query(map[string]any{"zpw_ver": 645, "zpw_type": g.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return g.parseGroupResult(data)
}

func (g *GroupAPI) ChangeGroupOwner(newAdminID string, groupID string) (*worker.Group, error) {
	enc, err := g.Encode(map[string]any{"grid": groupID, "newAdminId": newAdminID, "imei": g.IMEIValue(), "language": "vi"})
	if err != nil {
		return nil, err
	}
	data, err := g.GetJSON("https://tt-group-wpa.chat.zalo.me/api/group/change-owner", g.Query(map[string]any{"params": enc, "zpw_ver": 645, "zpw_type": g.APILoginType}))
	if err != nil {
		return nil, err
	}
	return g.parseGroupResult(data)
}

func (g *GroupAPI) ChangeGroupSetting(groupID string, defaultMode string, kwargs map[string]any) (*worker.Group, error) {
	if kwargs == nil {
		kwargs = map[string]any{}
	}
	defSetting := g.currentGroupSettings(groupID, defaultMode)
	payload := map[string]any{
		"blockName":        util.AsInt(firstNonNil(kwargs["blockName"], defSetting["blockName"], 1)),
		"signAdminMsg":     util.AsInt(firstNonNil(kwargs["signAdminMsg"], defSetting["signAdminMsg"], 1)),
		"addMemberOnly":    util.AsInt(firstNonNil(kwargs["addMemberOnly"], defSetting["addMemberOnly"], 0)),
		"setTopicOnly":     util.AsInt(firstNonNil(kwargs["setTopicOnly"], defSetting["setTopicOnly"], 1)),
		"enableMsgHistory": util.AsInt(firstNonNil(kwargs["enableMsgHistory"], defSetting["enableMsgHistory"], 1)),
		"lockCreatePost":   util.AsInt(firstNonNil(kwargs["lockCreatePost"], defSetting["lockCreatePost"], 1)),
		"lockCreatePoll":   util.AsInt(firstNonNil(kwargs["lockCreatePoll"], defSetting["lockCreatePoll"], 1)),
		"joinAppr":         util.AsInt(firstNonNil(kwargs["joinAppr"], defSetting["joinAppr"], 1)),
		"bannFeature":      util.AsInt(firstNonNil(kwargs["bannFeature"], defSetting["bannFeature"], 0)),
		"dirtyMedia":       util.AsInt(firstNonNil(kwargs["dirtyMedia"], defSetting["dirtyMedia"], 0)),
		"banDuration":      util.AsInt(firstNonNil(kwargs["banDuration"], defSetting["banDuration"], 0)),
		"lockSendMsg":      util.AsInt(firstNonNil(kwargs["lockSendMsg"], defSetting["lockSendMsg"], 0)),
		"lockViewMember":   util.AsInt(firstNonNil(kwargs["lockViewMember"], defSetting["lockViewMember"], 0)),
		"blocked_members":  firstNonNil(kwargs["blocked_members"], []any{}),
		"grid":             groupID, "imei": g.IMEIValue(),
	}
	enc, err := g.Encode(payload)
	if err != nil {
		return nil, err
	}
	data, err := g.GetJSON("https://tt-group-wpa.chat.zalo.me/api/group/setting/update", g.Query(map[string]any{"params": enc, "zpw_ver": 645, "zpw_type": g.APILoginType}))
	if err != nil {
		return nil, err
	}
	return g.parseGroupResult(data)
}

func (g *GroupAPI) CheckGroup(link string) map[string]any {
	enc, err := g.Encode(map[string]any{"link": link, "avatar_size": 120, "member_avatar_size": 120, "mpage": 1})
	if err != nil {
		return map[string]any{"error_code": 1337, "error_message": err.Error()}
	}
	data, err := g.GetJSON("https://tt-group-wpa.chat.zalo.me/api/group/link/ginfo", g.Query(map[string]any{"params": enc, "zpw_ver": 650, "zpw_type": g.APILoginType}))
	if err != nil {
		return map[string]any{"error_code": 1337, "error_message": err.Error()}
	}
	if util.AsInt(data["error_code"]) == 0 {
		decoded, derr := g.ParseRaw(data)
		if derr == nil {
			if m := util.AsMap(decoded); len(m) > 0 && m["data"] != nil {
				return util.AsMap(m["data"])
			}
			return util.AsMap(decoded)
		}
		return map[string]any{"error_code": 1337, "error_message": derr.Error()}
	}
	return map[string]any{"error_code": util.AsInt(data["error_code"]), "error_message": util.AsString(data["error_message"])}
}

func (g *GroupAPI) CreateGroup(name string, description string, members any, nameChanged int, createLink int) (any, error) {
	if name == "" {
		name = "Default Group Name"
	}
	mems := g.membersOf(members)
	memberTypes := make([]int, len(mems))
	for i := range memberTypes {
		memberTypes[i] = -1
	}
	enc, err := g.Encode(map[string]any{"clientId": util.Now(), "gname": name, "gdesc": description, "members": mems, "memberTypes": memberTypes, "nameChanged": map[bool]int{true: 1, false: 0}[name != ""], "createLink": createLink, "clientLang": "vi", "imei": g.IMEIValue(), "zsource": 601})
	if err != nil {
		return nil, err
	}
	data, err := g.GetJSON("https://tt-group-wpa.chat.zalo.me/api/group/create/v2", g.Query(map[string]any{"params": enc, "zpw_ver": 645, "zpw_type": g.APILoginType}))
	if err != nil {
		return nil, err
	}
	return g.ParseRaw(data)
}

func (g *GroupAPI) GetBlockedMembers(grid string, page int, count int) map[string]any {
	if page <= 0 {
		page = 1
	}
	if count <= 0 {
		count = 50
	}
	enc, err := g.Encode(map[string]any{"grid": grid, "page": page, "count": count})
	if err != nil {
		return map[string]any{"success": false, "error_message": err.Error()}
	}
	data, err := g.GetJSON("https://tt-group-wpa.chat.zalo.me/api/group/blockedmems/list", g.Query(map[string]any{"params": enc, "zpw_ver": 650, "zpw_type": g.APILoginType}))
	if err != nil {
		return map[string]any{"success": false, "error_message": err.Error()}
	}
	if util.AsInt(data["error_code"]) == 0 {
		decoded, derr := g.ParseRaw(data)
		if derr != nil {
			return map[string]any{"success": false, "error_message": derr.Error()}
		}
		return map[string]any{"success": true, "blocked_members": decoded}
	}
	return map[string]any{"success": false, "error_code": data["error_code"], "error_message": util.AsString(data["error_message"])}
}

func (g *GroupAPI) GetGroupLink(threadID string) map[string]any {
	enc, err := g.Encode(map[string]any{"grid": threadID, "imei": g.IMEIValue()})
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	data, err := g.GetJSON("https://tt-group-wpa.chat.zalo.me/api/group/link/detail", g.Query(map[string]any{"params": enc, "zpw_ver": 645, "zpw_type": g.APILoginType}))
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	if util.AsInt(data["error_code"]) != 0 {
		return map[string]any{"error": util.AsString(firstNonNil(data["error_message"], data["data"], data["error_code"]))}
	}
	decoded, derr := util.DecodeAPIData(data["data"], g.Secret())
	if derr != nil {
		return map[string]any{"error": derr.Error()}
	}
	decoded = decodeMaybeJSON(decoded)
	if m, ok := decoded.(map[string]any); ok {
		return m
	}
	return map[string]any{"error": "invalid response format"}
}

func (g *GroupAPI) GetLastMsgs() (*worker.User, error) {
	enc, err := g.Encode(map[string]any{"threadIdLocalMsgId": util.JSONString(map[string]any{}), "imei": g.IMEIValue()})
	if err != nil {
		return nil, err
	}
	data, err := g.GetJSON("https://tt-convers-wpa.chat.zalo.me/api/preloadconvers/get-last-msgs", g.Query(map[string]any{"zpw_ver": 645, "zpw_type": g.APILoginType, "params": enc}))
	if err != nil {
		return nil, err
	}
	decoded, err := g.ParseRaw(data)
	if err != nil {
		return nil, err
	}
	return worker.UserFromDict(util.AsMap(decoded)["data"]), nil
}

func (g *GroupAPI) GetQRLink(userID string) (any, error) {
	form, err := g.EncodedForm(map[string]any{"fids": []string{userID}})
	if err != nil {
		return nil, err
	}
	data, err := g.PostJSON("https://tt-friend-wpa.chat.zalo.me/api/friend/mget-qr", g.Query(map[string]any{"zpw_ver": 641, "zpw_type": g.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return g.ParseRaw(data)
}

func (g *GroupAPI) GetRecentGroup(groupID string) (*worker.Group, error) {
	enc, err := g.Encode(map[string]any{"groupId": groupID, "globalMsgId": 10000000000000000, "count": 50, "msgIds": []any{}, "imei": g.IMEIValue(), "src": 1})
	if err != nil {
		return nil, err
	}
	data, err := g.GetJSON("https://tt-group-cm.chat.zalo.me/api/cm/getrecentv2", g.Query(map[string]any{"params": enc, "zpw_ver": 645, "zpw_type": g.APILoginType, "nretry": 0}))
	if err != nil {
		return nil, err
	}
	decoded, err := g.ParseRaw(data)
	if err != nil {
		return nil, err
	}
	m := util.AsMap(decoded)
	if raw := util.AsString(m["data"]); raw != "" {
		var out any
		if err := json.Unmarshal([]byte(raw), &out); err == nil {
			return worker.GroupFromDict(out), nil
		}
	}
	return worker.GroupFromDict(decoded), nil
}

func (g *GroupAPI) UpgradeCommunity(groupID string) (*worker.Group, error) {
	enc, err := g.Encode(map[string]any{"grId": groupID, "language": "vi"})
	if err != nil {
		return nil, err
	}
	data, err := g.GetJSON("https://tt-group-wpa.chat.zalo.me/api/group/upgrade/community", g.Query(map[string]any{"params": enc, "zpw_ver": 655, "zpw_type": g.APILoginType}))
	if err != nil {
		return nil, err
	}
	decoded, err := g.ParseRaw(data)
	if err != nil {
		return nil, err
	}
	return worker.GroupFromDict(decoded), nil
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

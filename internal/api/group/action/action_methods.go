package action

import (
	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

func (a *ActionAPI) AddAdmins(members any, groupID string) (*worker.Group, error) {
	enc, err := a.Encode(map[string]any{"grid": groupID, "members": a.membersOf(members), "imei": a.IMEIValue()})
	if err != nil {
		return nil, err
	}
	data, err := a.GetJSON("https://tt-group-wpa.chat.zalo.me/api/group/admins/add", a.Query(map[string]any{"params": enc, "zpw_ver": 645, "zpw_type": a.APILoginType}))
	if err != nil {
		return nil, err
	}
	return a.parseGroup(data)
}

func (a *ActionAPI) BlockUsers(members any, groupID string) (*worker.Group, error) {
	enc, err := a.Encode(map[string]any{"grid": groupID, "members": a.membersOf(members)})
	if err != nil {
		return nil, err
	}
	data, err := a.GetJSON("https://tt-group-wpa.chat.zalo.me/api/group/blockedmems/add", a.Query(map[string]any{"params": enc, "zpw_ver": 645, "zpw_type": a.APILoginType}))
	if err != nil {
		return nil, err
	}
	return a.parseGroup(data)
}

func (a *ActionAPI) CreatePoll(question string, options any, groupID string, expiredTime int64, pinAct bool, multiChoices bool, allowAddNewOption bool, hideVotePreview bool, isAnonymous bool) (*worker.Group, error) {
	opt := a.membersOf(options)
	form, err := a.EncodedForm(map[string]any{"group_id": groupID, "question": question, "options": opt, "expired_time": expiredTime, "pinAct": pinAct, "allow_multi_choices": multiChoices, "allow_add_new_option": allowAddNewOption, "is_hide_vote_preview": hideVotePreview, "is_anonymous": isAnonymous, "poll_type": 0, "src": 1, "imei": a.IMEIValue()})
	if err != nil {
		return nil, err
	}
	data, err := a.PostJSON("https://tt-group-wpa.chat.zalo.me/api/poll/create", a.Query(map[string]any{"zpw_ver": 645, "zpw_type": a.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return a.parseGroup(data)
}

func (a *ActionAPI) JoinGroup(inviteURL string) (any, error) {
	form, err := a.EncodedForm(map[string]any{"link": inviteURL, "clientLang": "en"})
	if err != nil {
		return nil, err
	}
	data, err := a.PostJSON("https://tt-group-wpa.chat.zalo.me/api/group/link/join", a.Query(map[string]any{"zpw_ver": 648, "zpw_type": a.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return a.parseRaw(data)
}

func (a *ActionAPI) KickUsers(members any, groupID string) (*worker.Group, error) {
	form, err := a.EncodedForm(map[string]any{"grid": groupID, "members": a.membersOf(members)})
	if err != nil {
		return nil, err
	}
	data, err := a.PostJSON("https://tt-group-wpa.chat.zalo.me/api/group/kickout", a.Query(map[string]any{"zpw_ver": 645, "zpw_type": a.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return a.parseGroup(data)
}

func (a *ActionAPI) LeaveGroup(groupID string, silent bool) (any, error) {
	form, err := a.EncodedForm(map[string]any{"grids": []string{groupID}, "imei": a.IMEIValue(), "silent": map[bool]int{true: 1, false: 0}[silent]})
	if err != nil {
		return nil, err
	}
	data, err := a.PostJSON("https://tt-group-wpa.chat.zalo.me/api/group/leave", a.Query(map[string]any{"zpw_ver": 648, "zpw_type": a.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return a.parseRaw(data)
}

func (a *ActionAPI) LockPoll(pollID int64) (*worker.Group, error) {
	form, err := a.EncodedForm(map[string]any{"poll_id": pollID, "imei": a.IMEIValue()})
	if err != nil {
		return nil, err
	}
	data, err := a.PostJSON("https://tt-group-wpa.chat.zalo.me/api/poll/end", a.Query(map[string]any{"zpw_ver": 645, "zpw_type": a.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return a.parseGroup(data)
}

func (a *ActionAPI) RemoveAdmins(members any, groupID string) (*worker.Group, error) {
	enc, err := a.Encode(map[string]any{"grid": groupID, "members": a.membersOf(members), "imei": a.IMEIValue()})
	if err != nil {
		return nil, err
	}
	data, err := a.GetJSON("https://tt-group-wpa.chat.zalo.me/api/group/admins/remove", a.Query(map[string]any{"params": enc, "zpw_ver": 645, "zpw_type": a.APILoginType}))
	if err != nil {
		return nil, err
	}
	return a.parseGroup(data)
}

func (a *ActionAPI) UnblockUsers(members any, groupID string) (*worker.Group, error) {
	enc, err := a.Encode(map[string]any{"grid": groupID, "members": a.membersOf(members)})
	if err != nil {
		return nil, err
	}
	data, err := a.GetJSON("https://tt-group-wpa.chat.zalo.me/api/group/blockedmems/remove", a.Query(map[string]any{"params": enc, "zpw_ver": 645, "zpw_type": a.APILoginType}))
	if err != nil {
		return nil, err
	}
	return a.parseGroup(data)
}

func (a *ActionAPI) VotePoll(pollID int64, optionIDs any, groupID string) (any, error) {
	ids := make([]int, 0)
	switch t := optionIDs.(type) {
	case []int:
		ids = append(ids, t...)
	case []any:
		for _, item := range t {
			ids = append(ids, util.AsInt(item))
		}
	default:
		ids = append(ids, util.AsInt(t))
	}
	payload := map[string]any{"poll_id": pollID, "option_ids": ids, "imei": a.IMEIValue()}
	if groupID != "" {
		payload["group_id"] = groupID
	}
	form, err := a.EncodedForm(payload)
	if err != nil {
		return nil, err
	}
	data, err := a.PostJSON("https://tt-group-wpa.chat.zalo.me/api/poll/vote", a.Query(map[string]any{"zpw_ver": 677, "zpw_type": a.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return a.parseRaw(data)
}

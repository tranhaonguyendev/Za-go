package get

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

func (g *GetAPI) GetGroupMember(threadID string) (any, error) {
	groupInfoRaw, err := g.FetchGroupInfo(threadID)
	if err != nil {
		return nil, err
	}

	groupInfo, ok := groupInfoRaw.(*worker.Group)
	if !ok {
		return nil, fmt.Errorf("group info invalid")
	}

	gridInfoMap, ok := groupInfo.Get("gridInfoMap").(map[string]any)
	if !ok {
		return nil, fmt.Errorf("gridInfoMap not found")
	}
	groupMeta, ok := gridInfoMap[threadID].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("group %s not found", threadID)
	}

	memList := toStringSlice(groupMeta["memVerList"])
	if len(memList) == 0 {
		return nil, fmt.Errorf("khong the lay thong tin thanh vien")
	}

	q := url.Values{}
	q.Set("zpw_ver", "649")
	q.Set("zpw_type", strconv.Itoa(g.APILoginType))

	profiles := map[string]any{}
	for _, chunk := range chunkSlice(memList, 500) {
		enc, err := util.ZaloEncode(map[string]any{
			"friend_pversion_map": chunk,
		}, g.State.GetSecretkey())
		if err != nil {
			return nil, err
		}

		form := url.Values{}
		form.Set("params", enc)

		resp, err := g.postJSON("https://tt-profile-wpa.chat.zalo.me/api/social/group/members?"+q.Encode(), form)
		if err != nil {
			return nil, err
		}
		decoded, err := g.decodeAPIData(resp)
		if err != nil {
			return nil, err
		}
		if m, ok := decoded.(map[string]any); ok {
			if p, ok := m["profiles"].(map[string]any); ok {
				for k, v := range p {
					profiles[k] = v
				}
			}
		}
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("khong the lay thong tin thanh vien")
	}
	return worker.GroupFromDict(map[string]any{"profiles": profiles}), nil
}

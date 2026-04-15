package group

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	base "github.com/tranhaonguyendev/za-go/internal/api/common"
	getapi "github.com/tranhaonguyendev/za-go/internal/api/get"
	"github.com/tranhaonguyendev/za-go/internal/app"
	"github.com/tranhaonguyendev/za-go/internal/util"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

type GroupAPI struct {
	*base.BaseAPI
	Getter *getapi.GetAPI
}

func NewGroupAPI(state *app.State, loginType int, hub *worker.Hub, getter *getapi.GetAPI) *GroupAPI {
	return &GroupAPI{BaseAPI: base.NewBaseAPI(state, loginType, hub), Getter: getter}
}

func (g *GroupAPI) parseGroupResult(data map[string]any) (*worker.Group, error) {
	return g.ParseGroup(data)
}

func (g *GroupAPI) membersOf(v any) []string {
	switch t := v.(type) {
	case []string:
		out := make([]string, 0, len(t))
		for _, item := range t {
			item = strings.TrimSpace(item)
			if item != "" {
				out = append(out, item)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			s := strings.TrimSpace(fmt.Sprintf("%v", item))
			if s != "" && s != "<nil>" {
				out = append(out, s)
			}
		}
		return out
	default:
		s := strings.TrimSpace(fmt.Sprintf("%v", v))
		if s == "" || s == "<nil>" {
			return []string{}
		}
		return []string{s}
	}
}

func (g *GroupAPI) antiRaidSettings() map[string]any {
	return map[string]any{
		"blockName": 1, "signAdminMsg": 1, "addMemberOnly": 0, "setTopicOnly": 1,
		"enableMsgHistory": 1, "lockCreatePost": 1, "lockCreatePoll": 1, "joinAppr": 1,
		"bannFeature": 0, "dirtyMedia": 0, "banDuration": 0, "lockSendMsg": 0, "lockViewMember": 0,
	}
}

func (g *GroupAPI) currentGroupSettings(groupID string, defaultMode string) map[string]any {
	if defaultMode == "anti-raid" || g.Getter == nil {
		return g.antiRaidSettings()
	}
	info, err := g.Getter.FetchGroupInfo(groupID)
	if err != nil {
		return map[string]any{}
	}
	obj, ok := info.(*worker.Group)
	if !ok {
		return map[string]any{}
	}
	gridInfoMap := util.AsMap(obj.Get("gridInfoMap"))
	grid := util.AsMap(gridInfoMap[groupID])
	return util.AsMap(grid["setting"])
}

func decodeMaybeJSON(v any) any {
	if s, ok := v.(string); ok {
		var out any
		if err := json.Unmarshal([]byte(s), &out); err == nil {
			return out
		}
	}
	return v
}

func uploadFile(path string) ([]byte, string, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	return buf, util.AsString(strings.TrimSpace(path[strings.LastIndex(path, "/")+1:])), nil
}

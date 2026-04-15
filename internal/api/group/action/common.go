package action

import (
	"fmt"
	"strings"

	base "github.com/nguyendev/zago/internal/api/common"
	"github.com/nguyendev/zago/internal/app"
	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

type ActionAPI struct{ *base.BaseAPI }

func NewActionAPI(state *app.State, loginType int, hub *worker.Hub) *ActionAPI {
	return &ActionAPI{BaseAPI: base.NewBaseAPI(state, loginType, hub)}
}

func (a *ActionAPI) membersOf(v any) []string {
	switch t := v.(type) {
	case []string:
		out := make([]string, 0, len(t))
		for _, item := range t {
			if s := strings.TrimSpace(item); s != "" {
				out = append(out, s)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			if s := strings.TrimSpace(fmt.Sprintf("%v", item)); s != "" && s != "<nil>" {
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

func (a *ActionAPI) parseGroup(data map[string]any) (*worker.Group, error) { return a.ParseGroup(data) }
func (a *ActionAPI) parseRaw(data map[string]any) (any, error)             { return a.ParseRaw(data) }
func (a *ActionAPI) form(payload map[string]any) (map[string]any, error)   { return payload, nil }

var _ = util.Now

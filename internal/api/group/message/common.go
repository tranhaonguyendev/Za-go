package message

import (
	"encoding/json"
	base "github.com/nguyendev/zago/internal/api/common"
	"github.com/nguyendev/zago/internal/app"
	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

type MessageAPI struct{ *base.BaseAPI }

func NewMessageAPI(state *app.State, loginType int, hub *worker.Hub) *MessageAPI {
	return &MessageAPI{BaseAPI: base.NewBaseAPI(state, loginType, hub)}
}

func (m *MessageAPI) parseGroup(data map[string]any) (*worker.Group, error) {
	return m.ParseGroup(data)
}

func jsonMap(raw string) map[string]any {
	out := map[string]any{}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

var _ = util.Now

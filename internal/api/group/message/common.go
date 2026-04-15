package message

import (
	"encoding/json"
	base "github.com/tranhaonguyendev/za-go/internal/api/common"
	"github.com/tranhaonguyendev/za-go/internal/app"
	"github.com/tranhaonguyendev/za-go/internal/util"
	"github.com/tranhaonguyendev/za-go/internal/worker"
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

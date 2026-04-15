package status

import (
	base "github.com/nguyendev/zago/internal/api/common"
	"github.com/nguyendev/zago/internal/app"
	"github.com/nguyendev/zago/internal/worker"
)

type StatusAPI struct{ *base.BaseAPI }

func NewStatusAPI(state *app.State, loginType int, hub *worker.Hub) *StatusAPI {
	return &StatusAPI{BaseAPI: base.NewBaseAPI(state, loginType, hub)}
}

func (s *StatusAPI) parseGroup(data map[string]any) (*worker.Group, error) { return s.ParseGroup(data) }

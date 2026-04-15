package status

import (
	base "github.com/tranhaonguyendev/za-go/internal/api/common"
	"github.com/tranhaonguyendev/za-go/internal/app"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

type StatusAPI struct{ *base.BaseAPI }

func NewStatusAPI(state *app.State, loginType int, hub *worker.Hub) *StatusAPI {
	return &StatusAPI{BaseAPI: base.NewBaseAPI(state, loginType, hub)}
}

func (s *StatusAPI) parseGroup(data map[string]any) (*worker.Group, error) { return s.ParseGroup(data) }

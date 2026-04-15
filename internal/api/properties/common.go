package properties

import (
	base "github.com/nguyendev/zago/internal/api/common"
	"github.com/nguyendev/zago/internal/app"
	"github.com/nguyendev/zago/internal/worker"
)

type PropertiesAPI struct {
	*base.BaseAPI
}

func NewPropertiesAPI(state *app.State, loginType int, hub *worker.Hub) *PropertiesAPI {
	return &PropertiesAPI{BaseAPI: base.NewBaseAPI(state, loginType, hub)}
}

package properties

import (
	base "github.com/tranhaonguyendev/za-go/internal/api/common"
	"github.com/tranhaonguyendev/za-go/internal/app"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

type PropertiesAPI struct {
	*base.BaseAPI
}

func NewPropertiesAPI(state *app.State, loginType int, hub *worker.Hub) *PropertiesAPI {
	return &PropertiesAPI{BaseAPI: base.NewBaseAPI(state, loginType, hub)}
}

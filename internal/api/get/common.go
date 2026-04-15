package get

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	base "github.com/nguyendev/zago/internal/api/common"
	"github.com/nguyendev/zago/internal/app"
	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

type GetAPI struct {
	*base.BaseAPI
}

func NewGetAPI(state *app.State, loginType int, hub *worker.Hub) *GetAPI {
	return &GetAPI{BaseAPI: base.NewBaseAPI(state, loginType, hub)}
}

func (g *GetAPI) decodeAPIData(resp map[string]any) (any, error) {
	return util.ParseResponseEnvelope(resp, g.State.GetSecretkey())
}

func (g *GetAPI) getJSON(rawURL string) (map[string]any, error) {
	resp, err := g.State.GetSession(rawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	out := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (g *GetAPI) postJSON(rawURL string, form url.Values) (map[string]any, error) {
	resp, err := g.State.PostSession(rawURL, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	out := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func toStringSlice(v any) []string {
	out := []string{}
	switch t := v.(type) {
	case []any:
		for _, it := range t {
			s := strings.TrimSpace(fmt.Sprintf("%v", it))
			if s != "" && s != "<nil>" {
				out = append(out, s)
			}
		}
	case []string:
		for _, s := range t {
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}

func chunkSlice(in []string, size int) [][]string {
	if size <= 0 {
		size = 500
	}
	out := [][]string{}
	for i := 0; i < len(in); i += size {
		j := i + size
		if j > len(in) {
			j = len(in)
		}
		out = append(out, in[i:j])
	}
	return out
}

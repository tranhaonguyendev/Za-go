package properties

import (
	"os"
	"time"

	"github.com/tranhaonguyendev/za-go/internal/app"
	"github.com/tranhaonguyendev/za-go/internal/util"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

func (p *PropertiesAPI) ChangeAccountAvatar(filePath string, width int, height int, language string, size int64) (*worker.User, error) {
	if width == 0 {
		width = 500
	}
	if height == 0 {
		height = 500
	}
	if language == "" {
		language = "vn"
	}
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if size == 0 {
		size = int64(len(buf))
	}
	meta := map[string]any{
		"origin":    map[string]any{"width": width, "height": height},
		"processed": map[string]any{"width": width, "height": height, "size": size},
	}
	enc, err := p.Encode(map[string]any{
		"avatarSize": 120,
		"clientId":   p.UIDValue() + util.FormatTime("%H:%M %d/%m/%Y"),
		"language":   language,
		"metaData":   util.JSONString(meta),
	})
	if err != nil {
		return nil, err
	}
	q := p.Query(map[string]any{"zpw_ver": 645, "zpw_type": p.APILoginType, "params": enc})
	files := []app.MultipartFile{{FieldName: "fileContent", FileName: p.GetFileName(filePath, "avatar"), Content: buf, ContentType: "application/octet-stream"}}
	data, err := p.PostMultipartJSON("https://tt-files-wpa.chat.zalo.me/api/profile/upavatar", q, nil, files, 20*time.Second)
	if err != nil {
		return nil, err
	}
	return p.ParseUser(data)
}

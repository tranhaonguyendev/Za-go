package host

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/tranhaonguyendev/za-go/internal/app"
	core "github.com/tranhaonguyendev/za-go/internal/core"
	"github.com/tranhaonguyendev/za-go/internal/util"
)

func (u *UploadAPI) UploadImage(filePath string, threadID string, threadType core.ThreadType) (map[string]any, error) {
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileName := filepath.Base(filePath)
	if fileName == "" {
		fileName = "image"
	}
	payload := map[string]any{
		"totalChunk": 1,
		"fileName":   fileName,
		"clientId":   util.Now(),
		"totalSize":  len(buf),
		"imei":       u.IMEIValue(),
		"isE2EE":     0,
		"jxl":        0,
		"chunkId":    1,
	}
	query := u.Query(map[string]any{"zpw_ver": 645, "zpw_type": u.APILoginType})
	baseURL := "https://tt-files-wpa.chat.zalo.me/api/"
	switch threadType {
	case core.USER:
		baseURL += "message/photo_original/upload"
		query.Set("type", "2")
		payload["toid"] = threadID
	case core.GROUP:
		baseURL += "group/photo_original/upload"
		query.Set("type", "11")
		payload["grid"] = threadID
	default:
		return nil, fmt.Errorf("thread type is invalid")
	}
	enc, err := u.Encode(payload)
	if err != nil {
		return nil, err
	}
	query.Set("params", enc)
	files := []app.MultipartFile{{FieldName: "chunkContent", FileName: fileName, Content: buf, ContentType: "application/octet-stream"}}
	data, err := u.PostMultipartJSON(baseURL, query, nil, files, 15*time.Second)
	if err != nil {
		return nil, err
	}
	decoded, err := u.ParseRaw(data)
	if err != nil {
		return nil, err
	}
	if m, ok := decoded.(map[string]any); ok {
		if _, exists := m["clientFileId"]; !exists {
			m["clientFileId"] = strconv.FormatInt(util.Now()-1000, 10)
		}
		return m, nil
	}
	return util.AsMap(decoded), nil
}

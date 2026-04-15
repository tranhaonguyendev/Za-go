package host

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	base "github.com/nguyendev/zago/internal/api/common"
	"github.com/nguyendev/zago/internal/app"
	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/worker"
)

type UploadAPI struct {
	*base.BaseAPI
}

func NewUploadAPI(state *app.State, loginType int, hub *worker.Hub) *UploadAPI {
	return &UploadAPI{BaseAPI: base.NewBaseAPI(state, loginType, hub)}
}

func (u *UploadAPI) buildUploadBase(threadID string, threadType core.ThreadType) (string, map[string]any, map[string]any, error) {
	baseURL := "https://tt-files-wpa.chat.zalo.me/api/"
	switch threadType {
	case core.USER:
		return baseURL + "message/", map[string]any{"type": 2}, map[string]any{"toid": threadID}, nil
	case core.GROUP:
		return baseURL + "group/", map[string]any{"type": 11}, map[string]any{"grid": threadID}, nil
	default:
		return "", nil, nil, fmt.Errorf("thread type is invalid")
	}
}

func (u *UploadAPI) buildUploadURLs() map[string]string {
	return map[string]string{
		"image":  "photo_original/upload",
		"aac":    "voice/upload",
		"video":  "asyncfile/upload",
		"gif":    "gif?",
		"others": "asyncfile/upload",
	}
}

func (u *UploadAPI) pickFileType(ext string, fileSize int64, maxSize int64, maxTypeName string) (string, string) {
	e := strings.ToLower(ext)
	switch e {
	case "jpg", "jpeg", "png":
		return "image", ""
	case "mp3", "aac", "m4a", "flac":
		if fileSize > maxSize {
			return "others", maxTypeName
		}
		return "aac", ""
	case "mp4":
		return "video", ""
	default:
		return "others", ""
	}
}

func (u *UploadAPI) failPayload(filePath string, fileType string, reason string, maxType string) map[string]any {
	totalSize := int64(0)
	if st, err := os.Stat(filePath); err == nil {
		totalSize = st.Size()
	}
	return map[string]any{
		"ok":        false,
		"reason":    reason,
		"fileName":  filepath.Base(filePath),
		"totalSize": totalSize,
		"fileType":  fileType,
		"fileUrl":   "",
		"fileId":    "-1",
		"maxtype":   maxType,
	}
}

func (u *UploadAPI) waitUploadCallback(filePath, fileType, fileID, maxType string, timeout time.Duration) map[string]any {
	result := map[string]any{
		"fileName":  filepath.Base(filePath),
		"totalSize": u.GetLocalSize(filePath),
		"fileType":  fileType,
		"fileUrl":   "",
		"fileId":    fileID,
	}
	if u.Hub == nil {
		result["ok"] = false
		return result
	}
	ch := u.Hub.RegisterUploadWaiter(fileID)
	defer u.Hub.CancelUploadWaiter(fileID)
	select {
	case evt, ok := <-ch:
		if ok {
			result["fileUrl"] = evt.FileURL
			result["fileId"] = evt.FileID
		}
	case <-time.After(timeout):
	}
	if maxType != "" && result["fileUrl"] != "" {
		result["fileUrl"] = fmt.Sprintf("%v/%s", result["fileUrl"], maxType)
	}
	result["ok"] = result["fileUrl"] != ""
	return result
}

func chunkCount(size int64, chunkSize int64) int {
	return int(math.Ceil(float64(size) / float64(chunkSize)))
}

func (u *UploadAPI) selectChunkSize(fileSize int64, fileType string) int64 {
	switch {
	case (fileType == "video" || fileType == "others") && fileSize >= 256*1024*1024:
		return 16 * 1024 * 1024
	case (fileType == "video" || fileType == "others") && fileSize >= 64*1024*1024:
		return 8 * 1024 * 1024
	default:
		return 3 * 1024 * 1024
	}
}

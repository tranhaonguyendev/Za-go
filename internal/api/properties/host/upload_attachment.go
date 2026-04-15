package host

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nguyendev/zago/internal/app"
	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/util"
)

const (
	uploadRetryCount   = 5
	uploadWaitTimeout  = 5 * time.Second
	uploadPostTimeout  = 10 * time.Second
	uploadChunkSize    = int64(3145728)
	uploadMaxSizeBytes = int64(9 * 1000 * 1000)
)

func (u *UploadAPI) UploadAttachment(filePath string, threadID string, threadType core.ThreadType) (map[string]any, error) {
	st, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	fileSize := st.Size()
	ext := u.GetExt(filePath)
	maxTypeName := strings.ReplaceAll(u.UIDValue(), " ", "-") + ".aac"
	fileType, maxType := u.pickFileType(ext, fileSize, uploadMaxSizeBytes, maxTypeName)
	baseURL, baseQueryMap, extraParams, err := u.buildUploadBase(threadID, threadType)
	if err != nil {
		return nil, err
	}
	clientID := util.Now()
	totalChunks := chunkCount(fileSize, uploadChunkSize)
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	urls := u.buildUploadURLs()
	chunkBuf := make([]byte, uploadChunkSize)
	uploadedBytes := int64(0)
	fmt.Printf("Upload start: %s (%d bytes, %d chunks)\n", filepath.Base(filePath), fileSize, totalChunks)
	for chunkID := 1; chunkID <= totalChunks; chunkID++ {
		n, readErr := f.Read(chunkBuf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, chunkBuf[:n])
			payload := map[string]any{
				"totalChunk": totalChunks,
				"fileName":   filepath.Base(filePath),
				"fileType":   fileType,
				"clientId":   clientID,
				"totalSize":  fileSize,
				"imei":       u.IMEIValue(),
				"isE2EE":     0,
				"jxl":        0,
				"chunkId":    chunkID,
			}
			for k, v := range extraParams {
				payload[k] = v
			}
			enc, err := u.Encode(payload)
			if err != nil {
				return nil, err
			}
			query2 := u.Query(baseQueryMap)
			query2.Set("zpw_ver", "649")
			query2.Set("zpw_type", strconv.Itoa(u.APILoginType))
			query2.Set("params", enc)
			files := []app.MultipartFile{{FieldName: "chunkContent", FileName: filepath.Base(filePath), Content: chunk, ContentType: "application/octet-stream"}}
			var data map[string]any
			var lastErr error
			for i := 0; i < uploadRetryCount; i++ {
				data, lastErr = u.PostMultipartJSON(baseURL+urls[fileType], query2, nil, files, uploadPostTimeout)
				if lastErr == nil && util.AsInt(data["error_code"]) == 0 {
					break
				}
			}
			if lastErr != nil {
				fmt.Printf("\nUpload failed at chunk %d/%d: %v\n", chunkID, totalChunks, lastErr)
				return u.failPayload(filePath, fileType, lastErr.Error(), maxType), nil
			}
			uploadedBytes += int64(n)
			percent := float64(uploadedBytes) * 100 / float64(fileSize)
			fmt.Printf("\rUploading %s: chunk %d/%d (%.0f%%)", filepath.Base(filePath), chunkID, totalChunks, percent)
			decoded, err := u.ParseRaw(data)
			if err != nil {
				fmt.Printf("\nUpload decode failed at chunk %d/%d: %v\n", chunkID, totalChunks, err)
				continue
			}
			m := util.AsMap(decoded)
			if fileID := util.AsString(m["fileId"]); fileID != "" && fileID != "-1" {
				fmt.Printf("\nUpload complete, waiting file URL for fileId=%s\n", fileID)
				return u.waitUploadCallback(filePath, fileType, fileID, maxType, uploadWaitTimeout), nil
			}
			if photoID := util.AsString(m["photoId"]); photoID != "" && util.AsBool(m["finished"]) {
				fmt.Printf("\nUpload complete with photoId=%s\n", photoID)
				m["ok"] = true
				m["fileType"] = fileType
				return m, nil
			}
		}
		if readErr != nil {
			break
		}
	}
	fmt.Printf("\nUpload failed: no successful chunk\n")
	return u.failPayload(filePath, fileType, "no_successful_chunk", maxType), nil
}

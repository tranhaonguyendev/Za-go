package handle

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/util"
)

func (s *SendAPI) SendFile(fileURL string, threadID string, threadType core.ThreadType, fileName string, fileSize int, extension string, ttl int, localPath string) (any, error) {
	var content []byte
	if localPath != "" {
		if b, err := os.ReadFile(localPath); err == nil {
			content = b
			if fileSize == 0 {
				fileSize = len(content)
			}
		}
	} else if fileSize == 0 {
		b, err := s.remoteGetBytes(fileURL)
		if err != nil {
			return nil, err
		}
		content = b
		fileSize = len(content)
	}

	checksum := md5Hex(content)
	if checksum == "" {
		checksum = md5Hex([]byte{})
	}
	if parts := strings.Split(fileName, "."); len(parts) == 2 && parts[1] != "" {
		extension = parts[1]
	}
	if extension == "" {
		extension = "nullType"
	}

	params := url.Values{}
	params.Set("zpw_ver", "645")
	params.Set("zpw_type", strconv.Itoa(s.APILoginType))
	params.Set("nretry", "0")

	payload := map[string]any{
		"fileId":      strconv.FormatInt(util.Now()*2, 10),
		"checksum":    checksum,
		"checksumSha": "",
		"extension":   extension,
		"totalSize":   fileSize,
		"fileName":    defaultFileName(fileName, localPath),
		"clientId":    util.Now(),
		"fType":       1,
		"fileCount":   0,
		"fdata":       "{}",
		"fileUrl":     fileURL,
		"zsource":     401,
		"ttl":         ttl,
	}

	var endpoint string
	switch threadType {
	case core.USER:
		endpoint = "https://tt-files-wpa.chat.zalo.me/api/message/asyncfile/msg"
		payload["toid"] = threadID
		payload["imei"] = s.IMEI
	case core.GROUP:
		endpoint = "https://tt-files-wpa.chat.zalo.me/api/group/asyncfile/msg"
		payload["grid"] = threadID
	default:
		return nil, fmt.Errorf("thread type is invalid")
	}

	enc, err := util.ZaloEncode(payload, s.State.GetSecretkey())
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("params", enc)

	resp, err := s.State.PostSession(endpoint+"?"+params.Encode(), form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return s.parseThreadResponse(out, threadType, payload["clientId"])
}

func defaultFileName(fileName string, localPath string) string {
	if strings.TrimSpace(fileName) != "" {
		return fileName
	}
	if localPath != "" {
		base := filepath.Base(localPath)
		if base != "" {
			return base
		}
	}
	return "default"
}

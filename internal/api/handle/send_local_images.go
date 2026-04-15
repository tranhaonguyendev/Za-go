package handle

import (
	"fmt"
	"time"

	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

func (s *SendAPI) SendLocalImage(imagePath string, threadID string, threadType core.ThreadType, width int, height int, message *worker.Message, customPayload map[string]any, ttl int) (any, error) {
	if width == 0 {
		width = 2560
	}
	if height == 0 {
		height = 2560
	}
	payloadParams := map[string]any{}
	url := "https://tt-files-wpa.chat.zalo.me/api/message/photo_original/send"
	if customPayload != nil {
		if raw, ok := customPayload["params"].(map[string]any); ok {
			payloadParams = raw
		} else {
			return nil, fmt.Errorf("custom_payload is invalid")
		}
		if threadType == core.GROUP {
			url = "https://tt-files-wpa.chat.zalo.me/api/group/photo_original/send"
		}
	} else {
		uploadImage, err := s.Uploader.UploadImage(imagePath, threadID, threadType)
		if err != nil {
			return nil, err
		}
		desc := ""
		if message != nil {
			desc = message.Text
		}
		payloadParams = map[string]any{
			"photoId":   firstNonZero(uploadImage["photoId"], util.Now()*2),
			"clientId":  firstNonZero(uploadImage["clientFileId"], util.Now()-1000),
			"desc":      desc,
			"width":     width,
			"height":    height,
			"rawUrl":    uploadImage["normalUrl"],
			"thumbUrl":  uploadImage["thumbUrl"],
			"hdUrl":     uploadImage["hdUrl"],
			"thumbSize": "53932",
			"fileSize":  "247671",
			"hdSize":    "344622",
			"zsource":   -1,
			"jcp":       util.JSONString(map[string]any{"sendSource": 1, "convertible": "jxl"}),
			"ttl":       ttl,
			"imei":      s.IMEIValue(),
		}
		if message != nil && message.Mention != "" {
			payloadParams["mentionInfo"] = message.Mention
		}
		switch threadType {
		case core.USER:
			payloadParams["toid"] = threadID
			payloadParams["normalUrl"] = uploadImage["normalUrl"]
		case core.GROUP:
			url = "https://tt-files-wpa.chat.zalo.me/api/group/photo_original/send"
			payloadParams["grid"] = threadID
			payloadParams["oriUrl"] = uploadImage["normalUrl"]
		default:
			return nil, fmt.Errorf("thread type is invalid")
		}
	}
	form, err := s.EncodedForm(payloadParams)
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON(url, s.Query(map[string]any{"zpw_ver": 645, "zpw_type": s.APILoginType, "nretry": 0}), form)
	if err != nil {
		return nil, err
	}
	return s.ParseThread(data, threadType)
}

func (s *SendAPI) SendMultiImage(imageURLs []string, threadID string, threadType core.ThreadType, width int, height int, message *worker.Message, ttl int) ([]any, error) {
	if len(imageURLs) == 0 {
		return nil, fmt.Errorf("image url must be a list to be able to send multiple at once")
	}
	if width == 0 {
		width = 2560
	}
	if height == 0 {
		height = 2560
	}
	groupLayoutID := fmt.Sprintf("%d", time.Now().UnixNano())
	baseSeed := time.Now().UnixNano()
	out := make([]any, 0, len(imageURLs))
	for i, imageURL := range imageURLs {
		clientID := baseSeed + int64(i) + 1
		photoID := (clientID << 1) & 0x7FFFFFFFFFFFFFFF
		desc := ""
		if message != nil {
			desc = message.Text
		}
		payload := map[string]any{"photoId": photoID, "clientId": fmt.Sprintf("%d", clientID), "desc": desc, "width": width, "height": height, "groupLayoutId": groupLayoutID, "totalItemInGroup": len(imageURLs), "isGroupLayout": 1, "idInGroup": i, "rawUrl": imageURL, "thumbUrl": imageURL, "hdUrl": imageURL, "zsource": -1, "jcp": util.JSONString(map[string]any{"sendSource": 1, "convertible": "jxl"}), "ttl": ttl, "imei": s.IMEIValue()}
		url := "https://tt-files-wpa.chat.zalo.me/api/message/photo_original/send"
		if message != nil && message.Mention != "" {
			payload["mentionInfo"] = message.Mention
		}
		switch threadType {
		case core.USER:
			payload["toid"] = threadID
			payload["normalUrl"] = imageURL
		case core.GROUP:
			url = "https://tt-files-wpa.chat.zalo.me/api/group/photo_original/send"
			payload["grid"] = threadID
			payload["oriUrl"] = imageURL
		default:
			return nil, fmt.Errorf("thread type is invalid")
		}
		form := s.Query(map[string]any{"zpw_ver": 649, "zpw_type": s.APILoginType, "nretry": 0})
		enc, err := s.Encode(payload)
		if err != nil {
			return nil, err
		}
		form.Set("params", enc)
		data, err := s.PostJSON(url, nil, form)
		if err != nil {
			return nil, err
		}
		parsed, err := s.ParseStd(data, threadType, fmt.Sprintf("%d", clientID))
		if err != nil {
			return nil, err
		}
		out = append(out, parsed)
	}
	return out, nil
}

func (s *SendAPI) SendMultiLocalImage(imagePaths []string, threadID string, threadType core.ThreadType, width int, height int, message *worker.Message, ttl int) ([]any, error) {
	if len(imagePaths) == 0 {
		return nil, fmt.Errorf("image path must be a list to be able to send multiple at once")
	}
	if width == 0 {
		width = 2560
	}
	if height == 0 {
		height = 2560
	}
	groupLayoutID := fmt.Sprintf("%d", util.Now())
	out := make([]any, 0, len(imagePaths))
	for i, imagePath := range imagePaths {
		uploadImage, err := s.Uploader.UploadImage(imagePath, threadID, threadType)
		if err != nil {
			return nil, err
		}
		desc := ""
		if message != nil {
			desc = message.Text
		}
		payload := map[string]any{"params": map[string]any{"photoId": firstNonZero(uploadImage["photoId"], util.Now()*2), "clientId": firstNonZero(uploadImage["clientFileId"], util.Now()-1000), "desc": desc, "width": width, "height": height, "groupLayoutId": groupLayoutID, "totalItemInGroup": len(imagePaths), "isGroupLayout": 1, "idInGroup": i, "rawUrl": uploadImage["normalUrl"], "thumbUrl": uploadImage["thumbUrl"], "hdUrl": uploadImage["hdUrl"], "thumbSize": "53932", "fileSize": "247671", "hdSize": "344622", "zsource": -1, "jcp": util.JSONString(map[string]any{"sendSource": 1, "convertible": "jxl"}), "ttl": ttl, "imei": s.IMEIValue()}}
		if message != nil && message.Mention != "" {
			payload["params"].(map[string]any)["mentionInfo"] = message.Mention
		}
		switch threadType {
		case core.USER:
			payload["params"].(map[string]any)["toid"] = threadID
			payload["params"].(map[string]any)["normalUrl"] = uploadImage["normalUrl"]
		case core.GROUP:
			payload["params"].(map[string]any)["grid"] = threadID
			payload["params"].(map[string]any)["oriUrl"] = uploadImage["normalUrl"]
		default:
			return nil, fmt.Errorf("thread type is invalid")
		}
		r, err := s.SendLocalImage(imagePath, threadID, threadType, width, height, message, payload, ttl)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func firstNonZero(values ...any) any {
	for _, v := range values {
		if util.AsString(v) != "" && util.AsString(v) != "0" {
			return v
		}
	}
	return nil
}

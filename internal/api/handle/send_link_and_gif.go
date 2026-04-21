package handle

import (
	"fmt"
	"html"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tranhaonguyendev/za-go/internal/app"
	core "github.com/tranhaonguyendev/za-go/internal/core"
	"github.com/tranhaonguyendev/za-go/internal/util"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

func pickMetaTag(htmlText string, keys []string) string {
	for _, k := range keys {
		re1 := regexp.MustCompile(`(?is)<meta[^>]+(?:property|name)\s*=\s*["']` + regexp.QuoteMeta(k) + `["'][^>]+content\s*=\s*["']([^"']+)["']`)
		if m := re1.FindStringSubmatch(htmlText); len(m) > 1 {
			return strings.TrimSpace(html.UnescapeString(m[1]))
		}
		re2 := regexp.MustCompile(`(?is)<meta[^>]+content\s*=\s*["']([^"']+)["'][^>]+(?:property|name)\s*=\s*["']` + regexp.QuoteMeta(k) + `["']`)
		if m := re2.FindStringSubmatch(htmlText); len(m) > 1 {
			return strings.TrimSpace(html.UnescapeString(m[1]))
		}
	}
	return ""
}

func pickTitleTag(htmlText string) string {
	re := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	if m := re.FindStringSubmatch(htmlText); len(m) > 1 {
		return strings.TrimSpace(html.UnescapeString(m[1]))
	}
	return ""
}

func pickFavicon(htmlText string, baseURL string) string {
	re1 := regexp.MustCompile(`(?is)<link[^>]+rel=["'](?:shortcut icon|icon)["'][^>]+href=["']([^"']+)["']`)
	if m := re1.FindStringSubmatch(htmlText); len(m) > 1 {
		return resolveMaybeRelativeURL(baseURL, strings.TrimSpace(m[1]))
	}
	re2 := regexp.MustCompile(`(?is)<link[^>]+href=["']([^"']+)["'][^>]+rel=["'](?:shortcut icon|icon)["']`)
	if m := re2.FindStringSubmatch(htmlText); len(m) > 1 {
		return resolveMaybeRelativeURL(baseURL, strings.TrimSpace(m[1]))
	}
	if parsed, err := url.Parse(baseURL); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		return parsed.Scheme + "://" + parsed.Host + "/favicon.ico"
	}
	return ""
}

func resolveMaybeRelativeURL(baseURL string, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(raw), "http://") || strings.HasPrefix(strings.ToLower(raw), "https://") {
		return raw
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return raw
	}
	ref, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return base.ResolveReference(ref).String()
}

func (s *SendAPI) uploadThumbToZalo(thumbURL string) string {
	thumbURL = strings.TrimSpace(thumbURL)
	if thumbURL == "" {
		return ""
	}

	resp, err := s.State.GetSessionEx(thumbURL, nil, 15*time.Second, true, nil)
	if err != nil {
		return thumbURL
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil || len(buf) == 0 {
		return thumbURL
	}

	ctype := strings.ToLower(strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0]))
	if ctype == "" {
		ctype = "image/jpeg"
	}

	fileName := map[string]string{
		"image/jpeg": "thumb.jpg",
		"image/jpg":  "thumb.jpg",
		"image/png":  "thumb.png",
		"image/gif":  "thumb.gif",
		"image/webp": "thumb.webp",
	}[ctype]
	if fileName == "" {
		fileName = "thumb.jpg"
	}

	data, err := s.PostMultipartJSON(
		"https://tt-files-wpa.chat.zalo.me/api/message/photo_original/upload",
		s.Query(map[string]any{"zpw_ver": 645, "zpw_type": s.APILoginType}),
		nil,
		[]app.MultipartFile{{
			FieldName:   "fileContent",
			FileName:    fileName,
			Content:     buf,
			ContentType: ctype,
		}},
		15*time.Second,
	)
	if err != nil {
		return thumbURL
	}

	decoded, err := s.ParseRaw(data)
	if err != nil {
		return thumbURL
	}
	m := util.AsMap(decoded)
	for _, k := range []string{"normalUrl", "hdUrl", "thumb", "url"} {
		if v := strings.TrimSpace(util.AsString(m[k])); v != "" {
			return v
		}
	}
	return thumbURL
}

func (s *SendAPI) fetchLinkData(link string) (map[string]any, error) {
	u := strings.TrimSpace(link)
	if u == "" {
		return nil, fmt.Errorf("invalid link")
	}
	if !strings.HasPrefix(strings.ToLower(u), "http://") && !strings.HasPrefix(strings.ToLower(u), "https://") {
		u = "https://" + u
	}
	resp, err := s.State.GetSessionEx(u, nil, 20*time.Second, true, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	finalURL := resp.Request.URL.String()
	host := resp.Request.URL.Hostname()
	ctype := strings.ToLower(resp.Header.Get("Content-Type"))
	if !strings.Contains(ctype, "text/html") && !strings.Contains(ctype, "application/xhtml") {
		return map[string]any{"href": finalURL, "src": host, "title": "", "desc": "", "thumb": "", "icon": ""}, nil
	}
	htmlText := string(body)
	title := pickMetaTag(htmlText, []string{"og:title", "twitter:title"})
	if title == "" {
		title = pickTitleTag(htmlText)
	}
	desc := pickMetaTag(htmlText, []string{"og:description", "twitter:description", "description"})
	thumb := pickMetaTag(htmlText, []string{"og:image", "twitter:image", "twitter:image:src"})
	if thumb != "" {
		thumb = resolveMaybeRelativeURL(finalURL, thumb)
	}
	href := pickMetaTag(htmlText, []string{"og:url"})
	if href == "" {
		href = finalURL
	}
	icon := pickFavicon(htmlText, finalURL)
	return map[string]any{"href": href, "src": host, "title": title, "desc": desc, "thumb": thumb, "icon": icon}, nil
}

func (s *SendAPI) SendLink(linkURL string, threadID string, threadType core.ThreadType, message *worker.Message, ttl int) (any, error) {
	linkData, err := s.fetchLinkData(linkURL)
	if err != nil {
		return nil, err
	}
	msg := ""
	if message != nil {
		msg = message.Text
	}

	customTitle := strings.TrimSpace(util.AsString(linkData["title"]))
	customDesc := strings.TrimSpace(util.AsString(linkData["desc"]))
	customSrc := strings.TrimSpace(util.AsString(linkData["src"]))
	customHref := strings.TrimSpace(util.AsString(linkData["href"]))
	customIcon := strings.TrimSpace(util.AsString(linkData["icon"]))
	customThumb := strings.TrimSpace(util.AsString(linkData["thumb"]))
	cdnThumb := ""
	if customThumb != "" {
		cdnThumb = s.uploadThumbToZalo(customThumb)
	}

	innerParams := map[string]any{
		"redirect_url":          "",
		"src":                   customSrc,
		"mediaTitle":            customTitle,
		"title":                 customTitle,
		"desc":                  customDesc,
		"streamUrl":             "",
		"type":                  12,
		"linkType":              12,
		"artist":                "",
		"count":                 "",
		"stream_icon":           customIcon,
		"mediaId":               "",
		"video_duration":        0,
		"arid":                  0,
		"href":                  customHref,
		"tType":                 1,
		"tWidth":                486,
		"tHeight":               256,
		"width":                 250,
		"height":                250,
		"thumb_renew":           cdnThumb,
		"local_path_thumb_link": cdnThumb,
		"thumb_src_type":        1,
		"link_sub_type":         1,
		"video_brain": map[string]any{
			"thumb": cdnThumb,
			"title": customTitle,
			"desc":  customDesc,
			"src":   customSrc,
			"href":  customHref,
			"icon":  customIcon,
		},
	}

	payload := map[string]any{
		"msg":         msg,
		"title":       customTitle,
		"description": customDesc,
		"href":        customHref,
		"thumb":       cdnThumb,
		"thumb_renew": cdnThumb,
		"thumbWidth":  486,
		"thumbHeight": 256,
		"icon":        customIcon,
		"src":         customSrc,
		"type":        12,
		"action":      "recommened.link",
		"params":      util.JSONString(innerParams),
		"ttl":         ttl,
		"clientId":    util.Now(),
	}
	if message != nil && message.Mention != "" {
		payload["mentionInfo"] = message.Mention
	} else if threadType == core.GROUP {
		payload["mentionInfo"] = "[]"
	}
	url := "https://tt-chat4-wpa.chat.zalo.me/api/message/link"
	switch threadType {
	case core.USER:
		payload["toId"] = threadID
	case core.GROUP:
		url = "https://tt-group-wpa.chat.zalo.me/api/group/sendlink"
		payload["imei"] = s.IMEIValue()
		payload["grid"] = threadID
		payload["visibility"] = 0
	default:
		return nil, fmt.Errorf("thread type is invalid")
	}
	form, err := s.EncodedForm(payload)
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON(url, s.Query(map[string]any{"zpw_ver": 645, "zpw_type": s.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return s.ParseThread(data, threadType)
}

func (s *SendAPI) SendLocalGif(gifPath string, thumbnailURL string, threadID string, threadType core.ThreadType, gifName string, width int, height int, ttl int) (any, error) {
	buf, err := os.ReadFile(gifPath)
	if err != nil {
		return nil, err
	}
	if gifName == "" {
		gifName = filepath.Base(gifPath)
		if gifName == "" {
			gifName = "gifBot.gif"
		}
	}
	payload := map[string]any{"clientId": util.AsString(util.Now()), "fileName": gifName, "totalSize": len(buf), "width": width, "height": height, "msg": "", "type": 1, "ttl": ttl, "thumb": thumbnailURL, "checksum": md5Hex(buf), "totalChunk": 1, "chunkId": 1}
	query := map[string]any{"zpw_ver": 645, "zpw_type": s.APILoginType, "type": 1}
	url := "https://tt-files-wpa.chat.zalo.me/api/message/gif"
	switch threadType {
	case core.USER:
		payload["toid"] = threadID
	case core.GROUP:
		url = "https://tt-files-wpa.chat.zalo.me/api/group/gif"
		payload["visibility"] = 0
		payload["grid"] = threadID
	default:
		return nil, fmt.Errorf("thread type is invalid")
	}
	enc, err := s.Encode(payload)
	if err != nil {
		return nil, err
	}
	q := s.Query(query)
	q.Set("params", enc)
	files := []app.MultipartFile{{FieldName: "chunkContent", FileName: gifName, Content: buf, ContentType: "application/octet-stream"}}
	data, err := s.PostMultipartJSON(url, q, nil, files, 15*time.Second)
	if err != nil {
		return nil, err
	}
	return s.ParseThread(data, threadType)
}

func (s *SendAPI) SendGifphy(gifPath string, thumbnailURL string, threadID string, threadType core.ThreadType, gifName string, width int, height int, ttl int) (any, error) {
	return s.SendLocalGif(gifPath, thumbnailURL, threadID, threadType, gifName, width, height, ttl)
}

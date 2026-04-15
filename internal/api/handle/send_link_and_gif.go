package handle

import (
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/nguyendev/zago/internal/app"
	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
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
		return map[string]any{"href": finalURL, "src": host, "title": "", "desc": finalURL, "thumb": "", "media": map[string]any{"type": 0, "count": 0, "mediaTitle": "", "artist": "", "streamUrl": "", "stream_icon": ""}}, nil
	}
	htmlText := string(body)
	title := pickMetaTag(htmlText, []string{"og:title", "twitter:title"})
	if title == "" {
		title = pickTitleTag(htmlText)
	}
	desc := pickMetaTag(htmlText, []string{"og:description", "twitter:description", "description"})
	if desc == "" {
		desc = finalURL
	}
	thumb := pickMetaTag(htmlText, []string{"og:image", "twitter:image", "twitter:image:src"})
	href := pickMetaTag(htmlText, []string{"og:url"})
	if href == "" {
		href = finalURL
	}
	return map[string]any{"href": href, "src": host, "title": title, "desc": desc, "thumb": thumb, "media": map[string]any{"type": 0, "count": 0, "mediaTitle": "", "artist": "", "streamUrl": "", "stream_icon": ""}}, nil
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
	payload := map[string]any{"msg": msg, "href": linkData["href"], "src": linkData["src"], "title": linkData["title"], "desc": linkData["desc"], "thumb": linkData["thumb"], "type": 0, "media": util.JSONString(map[string]any{"type": 0, "count": 0, "mediaTitle": "", "artist": "", "streamUrl": "", "stream_icon": ""}), "ttl": ttl, "clientId": util.Now()}
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

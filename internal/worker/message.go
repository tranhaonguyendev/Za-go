package worker

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	coreparse "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/util"
)

type Message struct {
	Text      string `json:"text"`
	Style     string `json:"style,omitempty"`
	Mention   string `json:"mention,omitempty"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func NewMessage(text string) Message { return Message{Text: text} }

func NewParsedMessage(text string, style string, mention string, parseMode string) (Message, error) {
	m := Message{Text: text, Style: style, Mention: mention, ParseMode: parseMode}
	if strings.TrimSpace(parseMode) == "" {
		return m, nil
	}
	if parseMode != "Markdown" && parseMode != "HTML" {
		return m, fmt.Errorf("invalid Parse Mode, only support Markdown and HTML")
	}
	plain, parsed := coreparse.Parse(text, parseMode)
	m.Text = plain
	baseStyles := make([]any, 0)
	if strings.TrimSpace(style) != "" {
		var parsedBase map[string]any
		if err := json.Unmarshal([]byte(style), &parsedBase); err == nil {
			if styles, ok := parsedBase["styles"].([]any); ok {
				baseStyles = append(baseStyles, styles...)
			}
		}
	}
	built := make([]any, 0, len(parsed)+1)
	for _, e := range parsed {
		built = append(built, MessageStyle(e.Start, e.Length, e.Type, e.Color, coreparse.ParseTextSize(e.Size), false))
	}
	addDefaultFont(built, utf8.RuneCountInString(plain), "10")
	if len(built) == 1 && len(baseStyles) == 0 {
		m.Style = util.JSONString(map[string]any{"styles": []any{built[0]}, "ver": 0})
	} else {
		m.Style = MultiMsgStyle(append(built, baseStyles...))
	}
	return m, nil
}

func addDefaultFont(styles []any, n int, size string) {
	type seg struct{ a, b int }
	fonts := make([]seg, 0)
	for _, item := range styles {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		st := util.AsString(m["st"])
		if strings.HasPrefix(st, "f_") {
			a := util.AsInt(m["start"])
			b := a + util.AsInt(m["len"])
			if b > a {
				fonts = append(fonts, seg{a: a, b: b})
			}
		}
	}
	if n <= 0 {
		return
	}
	if len(fonts) == 0 {
		styles = append(styles, MessageStyle(0, n, "font", "ffffff", size, false))
		return
	}
	sort.Slice(fonts, func(i, j int) bool { return fonts[i].a < fonts[j].a })
	merged := make([]seg, 0, len(fonts))
	cs, ce := fonts[0].a, fonts[0].b
	for _, cur := range fonts[1:] {
		if cur.a <= ce {
			if cur.b > ce {
				ce = cur.b
			}
		} else {
			merged = append(merged, seg{a: cs, b: ce})
			cs, ce = cur.a, cur.b
		}
	}
	merged = append(merged, seg{a: cs, b: ce})
	cur := 0
	for _, m := range merged {
		if m.a > cur {
			styles = append(styles, MessageStyle(cur, m.a-cur, "font", "ffffff", size, false))
		}
		if m.b > cur {
			cur = m.b
		}
	}
	if cur < n {
		styles = append(styles, MessageStyle(cur, n-cur, "font", "ffffff", size, false))
	}
}

func MessageStyle(offset, length int, style string, color string, size string, autoFormat bool) any {
	if color == "" {
		color = "ffffff"
	}
	if size == "" {
		size = "18"
	}
	styleMap := map[string]string{
		"bold":      "b",
		"italic":    "i",
		"underline": "u",
		"strike":    "s",
		"color":     "c_" + strings.ReplaceAll(color, "#", ""),
		"font":      "f_" + size,
	}
	st := styleMap[style]
	if st == "" {
		st = "f_18"
	}
	data := map[string]any{"start": offset, "len": length, "st": st}
	if !autoFormat {
		return data
	}
	return util.JSONString(map[string]any{"styles": []any{data}, "ver": 0})
}

func MultiMsgStyle(styles []any) string {
	return util.JSONString(map[string]any{"styles": styles, "ver": 0})
}

func MessageReaction(messageObject *MessageObject, autoFormat bool) any {
	msgID := util.AsInt(messageObject.Get("msgId"))
	cliID := util.AsInt(messageObject.Get("cliMsgId"))
	msgType := util.GetClientMessageType(util.AsString(messageObject.Get("msgType")))
	data := map[string]any{"gMsgID": msgID, "cMsgID": cliID, "msgType": msgType}
	if autoFormat {
		return []any{data}
	}
	return data
}

func Mention(uidOrList any, length int, offset int, autoFormat bool) any {
	dataList := buildMention(uidOrList, offset, length)
	if autoFormat {
		return util.JSONString(dataList)
	}
	return dataList
}

func buildMention(uidOrList any, offset int, length int) []map[string]any {
	switch t := uidOrList.(type) {
	case []any:
		out := make([]map[string]any, 0, len(t))
		for _, item := range t {
			out = append(out, buildMentionOne(item, offset, length)...)
		}
		return out
	case []string:
		out := make([]map[string]any, 0, len(t))
		for _, item := range t {
			out = append(out, buildMentionOne(item, offset, length)...)
		}
		return out
	default:
		return buildMentionOne(t, offset, length)
	}
}

func buildMentionOne(uid any, offset int, length int) []map[string]any {
	u := fmt.Sprintf("%v", uid)
	return []map[string]any{{
		"pos":  offset,
		"len":  length,
		"uid":  u,
		"type": map[bool]int{true: 1, false: 0}[u == "-1"],
	}}
}

func (m Message) String() string {
	b, _ := json.Marshal(m)
	return string(b)
}

func (m Message) WithMention(mention any) Message {
	m.Mention = util.AsString(mention)
	return m
}

func (m Message) WithStyle(style any) Message {
	m.Style = util.AsString(style)
	return m
}

func IntString(v int64) string { return strconv.FormatInt(v, 10) }

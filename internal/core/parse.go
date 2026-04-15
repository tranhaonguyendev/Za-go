package core

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type ParsedStyle struct {
	Start  int
	End    int
	Length int
	Type   string
	Color  string
	Size   string
}

func Parse(text string, parseMode string) (string, []ParsedStyle) {
	switch strings.TrimSpace(parseMode) {
	case "Markdown":
		return ParseMarkdown(text)
	case "HTML":
		return ParseHTML(text)
	default:
		return StripSimpleHTML(text)
	}
}

func ParseMarkdown(text string) (string, []ParsedStyle) {
	type tokenStyle struct {
		kind  string
		color string
	}

	tokenToType := map[string]tokenStyle{
		"**": {kind: "bold"},
		"__": {kind: "underline"},
		"~~": {kind: "strike"},
		"_":  {kind: "italic"},
		"==": {kind: "color", color: "#f7b503"},
		"++": {kind: "color", color: "#15a85f"},
		"!!": {kind: "color", color: "#db342e"},
	}
	sizeRe := regexp.MustCompile(`(?i)<textsize\s*=\s*(\d+)\s*>`)
	runes := []rune(text)
	n := len(runes)

	isWord := func(ch rune) bool {
		return ch != 0 && (unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_')
	}
	isSpace := func(ch rune) bool {
		return ch != 0 && unicode.IsSpace(ch)
	}
	canOpen := func(tok string, prev rune, next rune) bool {
		if next == 0 || isSpace(next) {
			return false
		}
		if tok == "_" {
			if isWord(prev) {
				return false
			}
			if !isWord(next) {
				return false
			}
			return true
		}
		return !isSpace(next)
	}
	canClose := func(tok string, prev rune, next rune) bool {
		if prev == 0 || isSpace(prev) {
			return false
		}
		if tok == "_" {
			if isWord(next) {
				return false
			}
			if !isWord(prev) {
				return false
			}
			return true
		}
		return true
	}
	hasValidClose := func(tok string, start int) bool {
		tokenRunes := []rune(tok)
		lt := len(tokenRunes)
		for j := start; j < n; j++ {
			idx := indexRunes(runes, tokenRunes, j)
			if idx == -1 {
				return false
			}
			var prev rune
			var next rune
			if idx-1 >= 0 {
				prev = runes[idx-1]
			}
			if idx+lt < n {
				next = runes[idx+lt]
			}
			if canClose(tok, prev, next) {
				return true
			}
			j = idx
		}
		return false
	}

	out := make([]rune, 0, n)
	elements := make([]ParsedStyle, 0)
	stack := map[string][]int{}
	sizeStack := map[string]int{}
	tokens := []string{"**", "__", "~~", "==", "++", "!!", "_"}

	for i := 0; i < n; {
		if m := sizeRe.FindStringSubmatchIndex(string(runes[i:])); m != nil && m[0] == 0 {
			match := string(runes[i:])[m[0]:m[1]]
			sub := sizeRe.FindStringSubmatch(match)
			if len(sub) > 1 {
				sz := sub[1]
				start, ok := sizeStack[sz]
				if ok {
					delete(sizeStack, sz)
					end := len(out)
					if end > start {
						elements = append(elements, ParsedStyle{Start: start, End: end, Length: end - start, Type: "font", Size: sz})
					}
				} else {
					sizeStack[sz] = len(out)
				}
			}
			i += utf8.RuneCountInString(match)
			continue
		}

		matched := ""
		for _, tok := range tokens {
			if hasPrefixRunes(runes, i, []rune(tok)) {
				matched = tok
				break
			}
		}
		if matched == "" {
			out = append(out, runes[i])
			i++
			continue
		}

		tokenRunes := []rune(matched)
		lt := len(tokenRunes)
		var prev rune
		var next rune
		if i-1 >= 0 {
			prev = runes[i-1]
		}
		if i+lt < n {
			next = runes[i+lt]
		}

		if st := stack[matched]; len(st) > 0 && canClose(matched, prev, next) {
			def := tokenToType[matched]
			start := st[len(st)-1]
			stack[matched] = st[:len(st)-1]
			end := len(out)
			if end > start {
				e := ParsedStyle{Start: start, End: end, Length: end - start, Type: def.kind}
				if def.color != "" {
					e.Color = def.color
				}
				elements = append(elements, e)
			}
			i += lt
			continue
		}

		if canOpen(matched, prev, next) && hasValidClose(matched, i+lt) {
			stack[matched] = append(stack[matched], len(out))
			i += lt
			continue
		}

		out = append(out, runes[i])
		i++
	}

	sort.Slice(elements, func(i, j int) bool { return elements[i].Start < elements[j].Start })
	return string(out), elements
}

func ParseHTML(text string) (string, []ParsedStyle) {
	tagToType := map[string]string{"b": "bold", "i": "italic", "u": "underline", "s": "strike"}
	colorMap := map[string]string{"red": "#db342e", "yellow": "#f7b503", "green": "#15a85f"}
	tagRe := regexp.MustCompile(`(?i)</?(b|i|u|s|red|yellow|green|textsize)(?:\s*=\s*(\d+))?>`)

	out := strings.Builder{}
	elements := make([]ParsedStyle, 0)
	stack := map[string][]int{}
	colorStack := map[string][]int{}
	sizeStack := map[string]int{}
	plainLen := 0
	pos := 0

	for _, m := range tagRe.FindAllStringSubmatchIndex(text, -1) {
		chunk := text[pos:m[0]]
		out.WriteString(chunk)
		plainLen += utf8.RuneCountInString(chunk)

		raw := text[m[0]:m[1]]
		tag := strings.ToLower(text[m[2]:m[3]])
		val := ""
		if len(m) >= 6 && m[4] != -1 {
			val = text[m[4]:m[5]]
		}
		isClose := strings.HasPrefix(raw, "</")

		if kind, ok := tagToType[tag]; ok {
			if isClose {
				if st := stack[tag]; len(st) > 0 {
					start := st[len(st)-1]
					stack[tag] = st[:len(st)-1]
					if plainLen > start {
						elements = append(elements, ParsedStyle{Start: start, End: plainLen, Length: plainLen - start, Type: kind})
					}
				}
			} else {
				stack[tag] = append(stack[tag], plainLen)
			}
		} else if color, ok := colorMap[tag]; ok {
			if isClose {
				if st := colorStack[tag]; len(st) > 0 {
					start := st[len(st)-1]
					colorStack[tag] = st[:len(st)-1]
					if plainLen > start {
						elements = append(elements, ParsedStyle{Start: start, End: plainLen, Length: plainLen - start, Type: "color", Color: color})
					}
				}
			} else {
				if st := colorStack[tag]; len(st) > 0 {
					start := st[len(st)-1]
					colorStack[tag] = st[:len(st)-1]
					if plainLen > start {
						elements = append(elements, ParsedStyle{Start: start, End: plainLen, Length: plainLen - start, Type: "color", Color: color})
					}
				} else {
					colorStack[tag] = append(colorStack[tag], plainLen)
				}
			}
		} else if tag == "textsize" {
			if isClose {
				if val != "" {
					if start, ok := sizeStack[val]; ok {
						delete(sizeStack, val)
						if plainLen > start {
							elements = append(elements, ParsedStyle{Start: start, End: plainLen, Length: plainLen - start, Type: "font", Size: val})
						}
					}
				}
			} else if val != "" {
				if start, ok := sizeStack[val]; ok {
					delete(sizeStack, val)
					if plainLen > start {
						elements = append(elements, ParsedStyle{Start: start, End: plainLen, Length: plainLen - start, Type: "font", Size: val})
					}
				} else {
					sizeStack[val] = plainLen
				}
			}
		}

		pos = m[1]
	}

	tail := text[pos:]
	out.WriteString(tail)
	plainLen += utf8.RuneCountInString(tail)

	sort.Slice(elements, func(i, j int) bool { return elements[i].Start < elements[j].Start })
	return out.String(), elements
}

func StripSimpleHTML(text string) (string, []ParsedStyle) {
	plain, _ := ParseHTML(text)
	return plain, []ParsedStyle{}
}

func ParseTextSize(size string) string {
	if _, err := strconv.Atoi(size); err == nil && strings.TrimSpace(size) != "" {
		return size
	}
	return "18"
}

func hasPrefixRunes(runes []rune, start int, token []rune) bool {
	if start < 0 || start+len(token) > len(runes) {
		return false
	}
	for i := range token {
		if runes[start+i] != token[i] {
			return false
		}
	}
	return true
}

func indexRunes(haystack []rune, needle []rune, start int) int {
	if len(needle) == 0 {
		return start
	}
	if start < 0 {
		start = 0
	}
	limit := len(haystack) - len(needle)
	for i := start; i <= limit; i++ {
		ok := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				ok = false
				break
			}
		}
		if ok {
			return i
		}
	}
	return -1
}

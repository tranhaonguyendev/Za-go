package util

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var HEADERS = map[string]string{
	"User-Agent":         "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Accept":             "application/json, text/plain, */*",
	"sec-ch-ua":          `"Not-A.Brand";v="99", "Chromium";v="124"`,
	"sec-ch-ua-mobile":   "?0",
	"sec-ch-ua-platform": `"Linux"`,
	"Origin":             "https://chat.zalo.me",
	"sec-fetch-site":     "same-site",
	"sec-fetch-mode":     "cors",
	"sec-fetch-dest":     "empty",
	"Referer":            "https://chat.zalo.me/",
	"Accept-Language":    "vi-VN,vi;q=0.9,en-US;q=0.6,en;q=0.5",
}

func Now() int64 { return time.Now().UnixMilli() }

func FormatTime(layout string, ms ...int64) string {
	t := time.Now()
	if len(ms) > 0 && ms[0] > 0 {
		t = time.UnixMilli(ms[0])
	}
	repl := strings.NewReplacer("%H", "15", "%M", "04", "%d", "02", "%m", "01", "%Y", "2006")
	return t.Format(repl.Replace(layout))
}

func DictToRawCookies(cookies map[string]string) string {
	if len(cookies) == 0 {
		return ""
	}
	parts := make([]string, 0, len(cookies))
	for k, v := range cookies {
		if strings.TrimSpace(k) == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, "; ")
}

func AsString(v any) string {
	return formatStringValue(v)
}

func formatStringValue(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case int:
		return strconv.Itoa(t)
	case int8:
		return strconv.FormatInt(int64(t), 10)
	case int16:
		return strconv.FormatInt(int64(t), 10)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	case int64:
		return strconv.FormatInt(t, 10)
	case uint:
		return strconv.FormatUint(uint64(t), 10)
	case uint8:
		return strconv.FormatUint(uint64(t), 10)
	case uint16:
		return strconv.FormatUint(uint64(t), 10)
	case uint32:
		return strconv.FormatUint(uint64(t), 10)
	case uint64:
		return strconv.FormatUint(t, 10)
	case float32:
		f := float64(t)
		if !math.IsNaN(f) && !math.IsInf(f, 0) && f == math.Trunc(f) {
			return strconv.FormatInt(int64(f), 10)
		}
		return strconv.FormatFloat(f, 'f', -1, 64)
	case float64:
		if !math.IsNaN(t) && !math.IsInf(t, 0) && t == math.Trunc(t) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return strings.TrimSpace(t.String())
	case []any:
		parts := make([]string, 0, len(t))
		for _, item := range t {
			parts = append(parts, formatStringValue(item))
		}
		return "[" + strings.Join(parts, " ") + "]"
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s:%s", k, formatStringValue(t[k])))
		}
		return "map[" + strings.Join(parts, " ") + "]"
	}
	return strings.TrimSpace(fmt.Sprintf("%v", v))
}

func AsInt(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case int8:
		return int(t)
	case int16:
		return int(t)
	case int32:
		return int(t)
	case int64:
		return int(t)
	case float32:
		return int(t)
	case float64:
		return int(t)
	case json.Number:
		if i, err := t.Int64(); err == nil {
			return int(i)
		}
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(t))
		return i
	}
	return 0
}

func AsInt64(v any) int64 {
	switch t := v.(type) {
	case int:
		return int64(t)
	case int8:
		return int64(t)
	case int16:
		return int64(t)
	case int32:
		return int64(t)
	case int64:
		return t
	case float32:
		return int64(t)
	case float64:
		return int64(t)
	case json.Number:
		i, _ := t.Int64()
		return i
	case string:
		i, _ := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		return i
	}
	return 0
}

func AsBool(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		s := strings.TrimSpace(strings.ToLower(t))
		return s == "1" || s == "true" || s == "yes"
	default:
		return AsInt(v) != 0
	}
}

func AsMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

func AsSlice(v any) []any {
	if s, ok := v.([]any); ok {
		return s
	}
	return []any{}
}

func JSONString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func DecodeJSONMap(data []byte) (map[string]any, error) {
	out := map[string]any{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func DecodeJSONAny(data []byte) (any, error) {
	var out any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func MD5Hex(content []byte) string {
	sum := md5.Sum(content)
	return hex.EncodeToString(sum[:])
}

func RandomInt() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var b strings.Builder
	for i := 0; i < 9; i++ {
		b.WriteByte(byte('0' + r.Intn(10)))
	}
	return b.String()
}

func GetHeader(buffer []byte) (byte, int, byte, error) {
	if len(buffer) < 4 {
		return 0, 0, 0, fmt.Errorf("invalid header")
	}
	return buffer[0], int(buffer[1]) | int(buffer[2])<<8, buffer[3], nil
}

func GetClientMessageType(msgType string) int {
	switch msgType {
	case "webchat":
		return 1
	case "chat.voice":
		return 31
	case "chat.photo":
		return 32
	case "chat.sticker":
		return 10
	case "chat.doodle":
		return 37
	case "chat.recommended", "chat.link":
		return 38
	case "chat.location.new":
		return 43
	case "chat.video.msg":
		return 44
	case "share.file":
		return 46
	case "chat.gif":
		return 49
	case "chat.webcontent":
		return 52
	case "chat.webcontent.v2":
		return 61
	default:
		return 1
	}
}

func GetGroupEventType(act string) string {
	switch act {
	case "join_request":
		return "join_request"
	case "join":
		return "join"
	case "leave":
		return "leave"
	case "remove_member":
		return "remove_member"
	case "block_member":
		return "block_member"
	case "update_setting":
		return "update_setting"
	case "update":
		return "update"
	case "new_link":
		return "new_link"
	case "add_admin":
		return "add_admin"
	case "remove_admin":
		return "remove_admin"
	default:
		return "unknown"
	}
}

func NormalizePhone(phone string) string {
	var b strings.Builder
	for _, ch := range strings.TrimSpace(phone) {
		if (ch >= '0' && ch <= '9') || ch == '+' {
			b.WriteRune(ch)
		}
	}
	s := b.String()
	s = strings.TrimPrefix(s, "+")
	if strings.HasPrefix(s, "84") {
		return s
	}
	if strings.HasPrefix(s, "0") {
		return "84" + s[1:]
	}
	return "84" + s
}

func ZaloEncode(params any, key string) (string, error) {
	rawKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", fmt.Errorf("decode key failed: %w", err)
	}
	b, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(rawKey)
	if err != nil {
		return "", err
	}
	b = pkcs7Pad(b, aes.BlockSize)
	iv := make([]byte, aes.BlockSize)
	enc := make([]byte, len(b))
	cbc := cipher.NewCBCEncrypter(block, iv)
	cbc.CryptBlocks(enc, b)
	return base64.StdEncoding.EncodeToString(enc), nil
}

func ZaloDecode(params string, key string) (map[string]any, error) {
	unq, err := url.PathUnescape(params)
	if err != nil {
		unq = params
	}
	rawKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("decode key failed: %w", err)
	}
	cipherText, err := base64.StdEncoding.DecodeString(unq)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(rawKey)
	if err != nil {
		return nil, err
	}
	if len(cipherText)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("invalid cipher block size")
	}
	iv := make([]byte, aes.BlockSize)
	plain := make([]byte, len(cipherText))
	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(plain, cipherText)
	plain, err = pkcs7Unpad(plain, aes.BlockSize)
	if err != nil {
		return nil, err
	}
	out, err := DecodeJSONMap(plain)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func ZWSDecode(parsed map[string]any, key string) (map[string]any, error) {
	payload := AsString(parsed["data"])
	encryptType := AsInt(parsed["encrypt"])
	if payload == "" || key == "" {
		return nil, nil
	}
	var decoded []byte
	switch encryptType {
	case 0:
		decoded = []byte(payload)
	case 1:
		decrypted, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return nil, err
		}
		gr, err := gzip.NewReader(bytes.NewReader(decrypted))
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		decoded, err = io.ReadAll(gr)
		if err != nil {
			return nil, err
		}
	case 2:
		unq, err := url.PathUnescape(payload)
		if err != nil {
			unq = payload
		}
		dataBytes, err := base64.StdEncoding.DecodeString(unq)
		if err != nil {
			return nil, err
		}
		if len(dataBytes) < 48 {
			return nil, fmt.Errorf("invalid encrypted ws packet")
		}
		iv := dataBytes[:16]
		additionalData := dataBytes[16:32]
		cipherSource := dataBytes[32:]
		rawKey, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			return nil, err
		}
		block, err := aes.NewCipher(rawKey)
		if err != nil {
			return nil, err
		}
		gcm, err := cipher.NewGCMWithNonceSize(block, len(iv))
		if err != nil {
			return nil, err
		}
		plain, err := gcm.Open(nil, iv, cipherSource, additionalData)
		if err != nil {
			return nil, err
		}
		gr, err := gzip.NewReader(bytes.NewReader(plain))
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		decoded, err = io.ReadAll(gr)
		if err != nil {
			return nil, err
		}
	default:
		return nil, nil
	}
	out, err := DecodeJSONMap(decoded)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func NormalizeDecodedData(v any) any {
	for {
		switch t := v.(type) {
		case map[string]any:
			if code, ok := t["error_code"]; ok && AsInt(code) == 0 {
				if d, exists := t["data"]; exists {
					v = d
					continue
				}
			}
			return t
		case string:
			txt := strings.TrimSpace(t)
			if txt == "" {
				return ""
			}
			if js, err := DecodeJSONAny([]byte(txt)); err == nil {
				v = js
				continue
			}
			return txt
		default:
			return v
		}
	}
}

func DecodeAPIData(raw any, secret string) (any, error) {
	if raw == nil {
		return nil, nil
	}
	switch v := raw.(type) {
	case map[string]any, []any:
		return NormalizeDecodedData(v), nil
	case string:
		txt := strings.TrimSpace(v)
		if txt == "" {
			return "", nil
		}
		if secret != "" {
			if decoded, err := ZaloDecode(txt, secret); err == nil {
				return NormalizeDecodedData(decoded), nil
			}
		}
		return NormalizeDecodedData(txt), nil
	default:
		return v, nil
	}
}

func ParseResponseEnvelope(resp map[string]any, secret string) (any, error) {
	if AsInt(resp["error_code"]) != 0 {
		msg := AsString(resp["error_message"])
		if msg == "" {
			msg = AsString(resp["data"])
		}
		return nil, fmt.Errorf("error #%d when sending requests: %s", AsInt(resp["error_code"]), msg)
	}
	decoded, err := DecodeAPIData(resp["data"], secret)
	if err != nil {
		return nil, err
	}
	if decoded == nil {
		return nil, fmt.Errorf("error #1337 when sending requests: data is nil")
	}
	return decoded, nil
}

func DecodeJSONString(v string) (any, error) {
	out, err := DecodeJSONAny([]byte(v))
	return out, err
}

func EnsureMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

func EnsureSlice(v any) []any {
	if s, ok := v.([]any); ok {
		return s
	}
	return []any{}
}

func DeepCopyMap(src map[string]any) map[string]any {
	if src == nil {
		return map[string]any{}
	}
	b, _ := json.Marshal(src)
	out, err := DecodeJSONMap(b)
	if err != nil {
		return map[string]any{}
	}
	return out
}

func pkcs7Pad(src []byte, blockSize int) []byte {
	pad := blockSize - (len(src) % blockSize)
	out := make([]byte, len(src)+pad)
	copy(out, src)
	for i := len(src); i < len(out); i++ {
		out[i] = byte(pad)
	}
	return out
}

func pkcs7Unpad(src []byte, blockSize int) ([]byte, error) {
	if len(src) == 0 || len(src)%blockSize != 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	pad := int(src[len(src)-1])
	if pad == 0 || pad > blockSize || pad > len(src) {
		return nil, fmt.Errorf("invalid padding")
	}
	for _, b := range src[len(src)-pad:] {
		if int(b) != pad {
			return nil, fmt.Errorf("invalid padding")
		}
	}
	return src[:len(src)-pad], nil
}

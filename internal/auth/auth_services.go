package auth

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tranhaonguyendev/za-go/internal/app"
	"github.com/tranhaonguyendev/za-go/internal/util"
)

type LoginAuth struct {
	State *app.State
	UID   string
}

func NewLoginAuth(state *app.State) *LoginAuth {
	return &LoginAuth{State: state}
}

func (a *LoginAuth) IsLoggedIn() bool {
	return a.State.IsLoggedin()
}

func (a *LoginAuth) GetSession() map[string]string {
	return a.State.GetCookies()
}

func (a *LoginAuth) SetSession(sessionCookies any) bool {
	d, err := cookiesToDict(sessionCookies)
	if err != nil || len(d) == 0 {
		return false
	}
	a.State.SetCookies(d)
	a.State.Config["raw_cookies"] = util.DictToRawCookies(d)
	if a.State.UserClientID != "" {
		a.UID = a.State.UserClientID
	}
	return true
}

func (a *LoginAuth) GetSecretKey() string {
	return a.State.GetSecretkey()
}

func (a *LoginAuth) SetSecretKey(secretkey string) bool {
	a.State.SetSecretkey(secretkey)
	return true
}

func (a *LoginAuth) GetSessionWsCookies() string {
	raw := util.DictToRawCookies(a.State.GetCookies())
	if raw != "" {
		return raw
	}
	if cfg, ok := a.State.Config["raw_cookies"]; ok {
		return strings.TrimSpace(fmt.Sprintf("%v", cfg))
	}
	return ""
}

func parseNetscape(text string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") && !strings.Contains(line, "\t") {
			continue
		}
		if strings.HasPrefix(line, "#HttpOnly_") {
			line = strings.TrimPrefix(line, "#HttpOnly_")
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 7 {
			continue
		}
		name := strings.TrimSpace(parts[5])
		value := strings.TrimSpace(parts[6])
		if name != "" {
			out[name] = value
		}
	}
	return out
}

func cookieStringToDict(s string) map[string]string {
	out := map[string]string{}
	for _, part := range strings.Split(s, ";") {
		part = strings.TrimSpace(part)
		if part == "" || !strings.Contains(part, "=") {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		if k != "" {
			out[k] = v
		}
	}
	return out
}

func cookiesToDict(cookies any) (map[string]string, error) {
	if cookies == nil {
		return map[string]string{}, nil
	}

	switch v := cookies.(type) {
	case map[string]string:
		out := map[string]string{}
		for k, val := range v {
			if strings.TrimSpace(k) != "" {
				out[k] = val
			}
		}
		return out, nil
	case map[string]any:
		for _, key := range []string{"sessionCookies", "cookie", "raw_cookies", "rawCookies"} {
			if nested, ok := v[key]; ok && nested != nil {
				return cookiesToDict(nested)
			}
		}
		if lst, ok := v["cookies"].([]any); ok {
			out := map[string]string{}
			for _, item := range lst {
				obj, ok := item.(map[string]any)
				if !ok {
					continue
				}
				name := strings.TrimSpace(fmt.Sprintf("%v", obj["name"]))
				val := strings.TrimSpace(fmt.Sprintf("%v", obj["value"]))
				if name != "" && val != "" {
					out[name] = val
				}
			}
			if len(out) > 0 {
				return out, nil
			}
		}
		out := map[string]string{}
		for k, val := range v {
			ks := strings.TrimSpace(k)
			if ks == "" || val == nil {
				continue
			}
			out[ks] = strings.TrimSpace(fmt.Sprintf("%v", val))
		}
		return out, nil
	case []any:
		out := map[string]string{}
		for _, item := range v {
			d, _ := cookiesToDict(item)
			for k, val := range d {
				out[k] = val
			}
		}
		return out, nil
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return map[string]string{}, nil
		}
		if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
			var parsed any
			if err := json.Unmarshal([]byte(s), &parsed); err == nil {
				return cookiesToDict(parsed)
			}
		}
		if strings.Contains(s, "Netscape HTTP Cookie File") || strings.Contains(s, "\tTRUE\t") || strings.Contains(s, "\tFALSE\t") {
			d := parseNetscape(s)
			if len(d) > 0 {
				return d, nil
			}
		}
		return cookieStringToDict(s), nil
	default:
		return nil, fmt.Errorf("unsupported cookie format")
	}
}

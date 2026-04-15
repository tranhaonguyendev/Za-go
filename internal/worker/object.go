package worker

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

type Object struct {
	data map[string]any
	kind string
}

func newObject(kind string, raw any) Object {
	m, ok := normalizeMap(raw).(map[string]any)
	if !ok || m == nil {
		m = map[string]any{}
	}
	return Object{data: m, kind: kind}
}

func normalizeMap(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = normalizeMap(val)
		}
		return out
	case []any:
		out := make([]any, 0, len(t))
		for _, item := range t {
			out = append(out, normalizeMap(item))
		}
		return out
	default:
		return v
	}
}

func (o Object) Get(key string) any {
	if o.data == nil {
		return nil
	}
	return o.data[key]
}

func (o Object) ToMap() map[string]any {
	out := map[string]any{}
	for k, v := range o.data {
		out[k] = normalizeMap(v)
	}
	return out
}

func (o Object) ToDict() map[string]any {
	return o.ToMap()
}

func (o Object) String() string {
	keys := make([]string, 0, len(o.data))
	for k := range o.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	attrs := make([]string, 0, len(keys))
	for _, k := range keys {
		attrs = append(attrs, fmt.Sprintf("%s=%s", k, formatValue(o.data[k])))
	}
	return fmt.Sprintf("%s(%s)", o.kind, strings.Join(attrs, ", "))
}

func formatValue(v any) string {
	switch t := v.(type) {
	case nil:
		return "<nil>"
	case float64:
		if t == math.Trunc(t) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case float32:
		f := float64(t)
		if f == math.Trunc(f) {
			return strconv.FormatInt(int64(f), 10)
		}
		return strconv.FormatFloat(f, 'f', -1, 32)
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
	case string:
		return fmt.Sprintf("%q", t)
	case []any:
		parts := make([]string, 0, len(t))
		for _, item := range t {
			parts = append(parts, formatValue(item))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s:%s", k, formatValue(t[k])))
		}
		return "map[" + strings.Join(parts, " ") + "]"
	default:
		return fmt.Sprintf("%v", t)
	}
}

type User struct{ Object }
type Group struct{ Object }
type ContextObject struct{ Object }
type MessageObject struct{ Object }
type EventObject struct{ Object }

func UserFromDict(raw any) *User {
	return &User{Object: newObject("User", raw)}
}

func GroupFromDict(raw any) *Group {
	return &Group{Object: newObject("Group", raw)}
}

func ContextFromDict(raw any) *ContextObject {
	return &ContextObject{Object: newObject("Context", raw)}
}

func MessageObjectFromDict(raw any) *MessageObject {
	return &MessageObject{Object: newObject("Message", raw)}
}

func EventObjectFromDict(raw any) *EventObject {
	return &EventObject{Object: newObject("GroupEvent", raw)}
}

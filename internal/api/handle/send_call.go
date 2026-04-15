package handle

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tranhaonguyendev/za-go/internal/util"
)

const groupCallName = "debug"

func (s *SendAPI) SendCall(targetID string, callID string, extras ...string) (map[string]any, error) {
	rtpAddress, rtcpAddress, codec := "", "", ""
	if len(extras) > 0 {
		rtpAddress = extras[0]
	}
	if len(extras) > 1 {
		rtcpAddress = extras[1]
	}
	if len(extras) > 2 {
		codec = extras[2]
	}
	if rtpAddress == "" {
		rtpAddress = "171.244.25.88:4601"
	}
	if rtcpAddress == "" {
		rtcpAddress = "171.244.25.88:4601"
	}
	if codec == "" {
		codec = `[{"dynamicFptime":0,"frmPtime":20,"name":"opus/16000/1","payload":112}]\n`
	}
	enc1, err := s.Encode(map[string]any{"calleeId": targetID, "callId": callID, "codec": "[]\n", "typeRequest": 1, "imei": s.IMEIValue()})
	if err != nil {
		return nil, err
	}
	data1, err := s.PostJSON(fmt.Sprintf("https://voicecall-wpa.chat.zalo.me/api/voicecall/requestcall?zpw_ver=646&zpw_type=%d", s.APILoginType), nil, s.Query(map[string]any{"params": enc1}))
	if err != nil {
		return nil, err
	}
	enc2, err := s.Encode(map[string]any{"calleeId": targetID, "rtcpAddress": rtcpAddress, "rtpAddress": rtpAddress, "codec": codec, "session": callID, "callId": callID, "imei": s.IMEIValue(), "subCommand": 3})
	if err != nil {
		return nil, err
	}
	data2, err := s.PostJSON(fmt.Sprintf("https://voicecall-wpa.chat.zalo.me/api/voicecall/request?zpw_ver=646&zpw_type=%d", s.APILoginType), nil, s.Query(map[string]any{"params": enc2}))
	if err != nil {
		return nil, err
	}
	if code := util.AsInt(data1["error_code"]); code != 0 {
		return nil, fmt.Errorf("error when sending call: %v | %v", data1, data2)
	}
	if code := util.AsInt(data2["error_code"]); code != 0 {
		return nil, fmt.Errorf("error when sending call: %v | %v", data1, data2)
	}
	return map[string]any{"requestcall": data1, "request": data2}, nil
}

func (s *SendAPI) CallGroupRequest(groupID string, userIDs []string, callID any, groupNames ...string) (map[string]any, error) {
	callID = resolveGroupCallID(callID)
	userIDs = cleanGroupCallUserIDs(userIDs)
	if len(userIDs) == 0 {
		return nil, fmt.Errorf("userIDs is empty")
	}
	groupName := resolveGroupCallName(groupNames...)
	form, err := s.EncodedForm(map[string]any{
		"groupId":     strings.TrimSpace(groupID),
		"callId":      callID,
		"typeRequest": 1,
		"data": util.JSONString(map[string]any{
			"extraData":   "",
			"groupAvatar": "",
			"groupId":     strings.TrimSpace(groupID),
			"groupName":   groupName,
			"maxUsers":    8,
			"noiseId":     userIDs,
		}),
		"partners": userIDs,
	})
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON(
		"https://voicecall-wpa.chat.zalo.me/api/voicecall/group/requestcall",
		s.Query(map[string]any{"zpw_ver": 667, "zpw_type": s.APILoginType}),
		form,
	)
	if err != nil {
		return nil, err
	}
	return s.decodeGroupCallResponse(data, true)
}

func (s *SendAPI) CallGroupAdd(userIDs []string, callID any, hostCall any, groupID string) (map[string]any, error) {
	callID = resolveGroupCallID(callID)
	userIDs = cleanGroupCallUserIDs(userIDs)
	if len(userIDs) == 0 {
		return nil, fmt.Errorf("userIDs is empty")
	}
	userID := userIDs[0]
	outerData := util.JSONString(map[string]any{
		"codec": "",
		"data": util.JSONString(map[string]any{
			"groupAvatar": "",
			"groupId":     groupCallScalar(groupID),
			"groupName":   groupCallName,
			"hostCall":    groupCallScalar(hostCall),
			"maxUsers":    8,
		}),
		"extendData":      "",
		"rtcpAddress":     "",
		"rtcpAddressIPv6": "",
		"rtpAddress":      "",
		"rtpAddressIPv6":  "",
	})
	form, err := s.EncodedForm(map[string]any{
		"callId":   callID,
		"callType": 1,
		"hostCall": groupCallScalar(hostCall),
		"data":     outerData,
		"session":  "",
		"partners": util.JSONString([]string{userID}),
		"groupId":  strings.TrimSpace(groupID),
	})
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON(
		"https://voicecall-wpa.chat.zalo.me/api/voicecall/group/adduser",
		s.Query(map[string]any{"zpw_ver": 667, "zpw_type": s.APILoginType}),
		form,
	)
	if err != nil {
		return nil, err
	}
	return s.decodeGroupCallResponse(data, false)
}

func (s *SendAPI) CallGroupCancel(callID any, hostCall any, groupID string) (map[string]any, error) {
	callID = resolveGroupCallID(callID)
	form, err := s.EncodedForm(map[string]any{
		"callId":   callID,
		"hostCall": groupCallScalar(hostCall),
		"data": util.JSONString(map[string]any{
			"callType":  1,
			"duration":  0,
			"extraData": "",
			"groupId":   groupCallScalar(groupID),
		}),
	})
	if err != nil {
		return nil, err
	}
	data, err := s.PostJSON(
		"https://voicecall-wpa.chat.zalo.me/api/voicecall/group/cancel",
		s.Query(map[string]any{"zpw_ver": 667, "zpw_type": s.APILoginType}),
		form,
	)
	if err != nil {
		return nil, err
	}
	return s.decodeGroupCallResponse(data, false)
}

func (s *SendAPI) CallGroup(groupID string, userIDs []string, extras ...any) (map[string]any, error) {
	callID := any(time.Now().Unix())
	if len(extras) > 0 {
		callID = resolveGroupCallID(extras[0])
	}
	userIDs = cleanGroupCallUserIDs(userIDs)
	if len(userIDs) == 0 {
		return nil, fmt.Errorf("userIDs is empty")
	}
	groupName := groupCallName
	if len(extras) > 1 {
		groupName = resolveGroupCallName(util.AsString(extras[1]))
	}

	requestData, err := s.CallGroupRequest(groupID, userIDs, callID, groupName)
	if err != nil {
		return nil, err
	}

	paramsData := parseGroupCallMap(requestData["params"])
	if paramsData == nil {
		return nil, fmt.Errorf("CallGroup failed: no params in response")
	}

	// Check API response status and error message
	if status := util.AsString(requestData["status"]); status == "2" {
		// Try to get error message from top level first, then from params
		msg := util.AsString(requestData["msg"])
		if msg == "" {
			msg = util.AsString(paramsData["msg"])
		}
		if msg != "" {
			return nil, fmt.Errorf("CallGroup failed: %s", msg)
		}
		return nil, fmt.Errorf("CallGroup failed: API returned status 2")
	}

	// Check for error message in params
	if msg := util.AsString(paramsData["msg"]); msg != "" && strings.Contains(strings.ToLower(msg), "does not support") {
		return nil, fmt.Errorf("CallGroup failed: %s", msg)
	}

	callSetting := parseGroupCallMap(paramsData["callSetting"])
	if callSetting == nil {
		return nil, fmt.Errorf("CallGroup failed: no callSetting in params")
	}

	session := util.AsString(callSetting["session"])
	servers := util.AsSlice(callSetting["servers"])
	if session == "" || len(servers) == 0 {
		errMsg := ""
		if msg := util.AsString(paramsData["msg"]); msg != "" {
			errMsg = fmt.Sprintf(" (%s)", msg)
		}
		return nil, fmt.Errorf("invalid call setup data: missing session or servers info%s", errMsg)
	}

	server := util.AsMap(servers[0])
	rtpAddress := util.AsString(server["rtpaddr"])
	rtcpAddress := util.AsString(server["rtcpaddr"])
	rtpAddressIPv6 := util.AsString(server["rtpaddrIPv6"])
	rtcpAddressIPv6 := util.AsString(server["rtcpaddrIPv6"])

	partnerIDs := groupCallStringSlice(requestData["partnerIds"])
	idcal := ""
	if len(partnerIDs) > 0 {
		idcal = partnerIDs[0]
	}

	maxUsers := util.AsInt(paramsData["maxUsers"])
	if maxUsers <= 0 {
		maxUsers = 8
	}

	callPayload := util.JSONString(map[string]any{
		"codec": "",
		"data": util.JSONString(map[string]any{
			"groupAvatar": "",
			"groupName":   groupName,
			"hostCall":    groupCallScalar(paramsData["hostCall"]),
			"maxUsers":    maxUsers,
			"noiseId":     groupCallMaybeOne(idcal),
		}),
		"extendData":      "",
		"rtcpAddress":     rtcpAddress,
		"rtcpAddressIPv6": rtcpAddressIPv6,
		"rtpAddress":      rtpAddress,
		"rtpAddressIPv6":  rtpAddressIPv6,
	})

	form, err := s.EncodedForm(map[string]any{
		"callId":   firstNonEmptyValue(paramsData["callId"], callID),
		"callType": 1,
		"data":     callPayload,
		"session":  session,
		"partners": util.JSONString(groupCallMaybeOne(idcal)),
		"groupId":  strings.TrimSpace(groupID),
	})
	if err != nil {
		return nil, err
	}

	data, err := s.PostJSON(
		"https://voicecall-wpa.chat.zalo.me/api/voicecall/group/request",
		s.Query(map[string]any{"zpw_ver": 667, "zpw_type": s.APILoginType}),
		form,
	)
	if err != nil {
		return nil, err
	}
	return s.decodeGroupCallResponse(data, false)
}

func (s *SendAPI) decodeGroupCallResponse(resp map[string]any, unwrapData bool) (map[string]any, error) {
	decoded, err := util.ParseResponseEnvelope(resp, s.State.GetSecretkey())
	if err != nil {
		return nil, err
	}
	if out, ok := parseGroupCallMapOK(decoded); ok {
		if unwrapData {
			if inner, ok := parseGroupCallMapOK(out["data"]); ok {
				return inner, nil
			}
		}
		return out, nil
	}
	return map[string]any{"data": decoded}, nil
}

func parseGroupCallMap(v any) map[string]any {
	out, _ := parseGroupCallMapOK(v)
	return out
}

func parseGroupCallMapOK(v any) (map[string]any, bool) {
	normalized := util.NormalizeDecodedData(v)
	if out, ok := normalized.(map[string]any); ok {
		return out, true
	}
	text, ok := normalized.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return nil, false
	}
	var parsed any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, false
	}
	out, ok := parsed.(map[string]any)
	return out, ok
}

func groupCallStringSlice(v any) []string {
	switch t := util.NormalizeDecodedData(v).(type) {
	case []string:
		return cleanGroupCallUserIDs(t)
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			text := strings.TrimSpace(util.AsString(item))
			if text != "" {
				out = append(out, text)
			}
		}
		return cleanGroupCallUserIDs(out)
	default:
		return nil
	}
}

func cleanGroupCallUserIDs(userIDs []string) []string {
	if len(userIDs) == 0 {
		return nil
	}
	out := make([]string, 0, len(userIDs))
	seen := make(map[string]struct{}, len(userIDs))
	for _, userID := range userIDs {
		userID = strings.TrimSpace(userID)
		if userID == "" {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		out = append(out, userID)
	}
	return out
}

func groupCallMaybeOne(userID string) []string {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return []string{}
	}
	return []string{userID}
}

func groupCallScalar(v any) any {
	switch t := v.(type) {
	case nil:
		return ""
	case int, int8, int16, int32, int64:
		return t
	case uint, uint8, uint16, uint32, uint64:
		return t
	case float32, float64:
		return t
	case json.Number:
		if i, err := t.Int64(); err == nil {
			return i
		}
		return t.String()
	case string:
		text := strings.TrimSpace(t)
		if text == "" {
			return ""
		}
		if i, err := strconv.ParseInt(text, 10, 64); err == nil {
			return i
		}
		return text
	default:
		text := strings.TrimSpace(util.AsString(v))
		if text == "" {
			return ""
		}
		if i, err := strconv.ParseInt(text, 10, 64); err == nil {
			return i
		}
		return text
	}
}

func resolveGroupCallID(v any) any {
	if v == nil {
		return time.Now().Unix()
	}
	switch t := v.(type) {
	case string:
		text := strings.TrimSpace(t)
		if text == "" {
			return time.Now().Unix()
		}
		return text
	default:
		if util.AsString(v) == "" {
			return time.Now().Unix()
		}
		return v
	}
}

func resolveGroupCallName(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return groupCallName
}

func firstNonEmptyValue(values ...any) any {
	for _, value := range values {
		if strings.TrimSpace(util.AsString(value)) != "" {
			return value
		}
	}
	return nil
}

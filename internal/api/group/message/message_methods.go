package message

import (
	"encoding/json"

	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

func (m *MessageAPI) DeleteMessage(msgID any, ownerID any, clientMsgID any, groupID string, onlyMe bool) (*worker.Group, error) {
	form, err := m.EncodedForm(map[string]any{"grid": groupID, "cliMsgId": util.Now(), "msgs": []any{map[string]any{"cliMsgId": util.AsString(clientMsgID), "globalMsgId": util.AsString(msgID), "ownerId": util.AsString(ownerID), "destId": groupID}}, "onlyMe": map[bool]int{true: 1, false: 0}[onlyMe]})
	if err != nil {
		return nil, err
	}
	data, err := m.PostJSON("https://tt-group-wpa.chat.zalo.me/api/group/deletemsg", m.Query(map[string]any{"zpw_ver": 645, "zpw_type": m.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return m.parseGroup(data)
}

func (m *MessageAPI) PinMessage(pinMsg *worker.MessageObject, groupID string) (*worker.Group, error) {
	content := util.AsMap(pinMsg.Get("content"))
	payload := map[string]any{"grid": groupID, "type": 2, "color": -14540254, "emoji": "📌", "startTime": -1, "duration": -1, "repeat": 0, "src": -1, "imei": m.IMEIValue(), "pinAct": 1}
	mt := util.AsString(pinMsg.Get("msgType"))
	base := map[string]any{"client_msg_id": util.AsString(pinMsg.Get("cliMsgId")), "global_msg_id": util.AsString(pinMsg.Get("msgId")), "senderUid": util.AsString(firstNonZero(pinMsg.Get("uidFrom"), m.UIDValue())), "senderName": util.AsString(pinMsg.Get("dName")), "msg_type": util.GetClientMessageType(mt)}
	switch mt {
	case "webchat", "chat.voice":
		if mt == "webchat" {
			base["title"] = util.AsString(pinMsg.Get("content"))
		}
	case "chat.photo", "chat.video.msg":
		base["thumb"] = util.AsString(content["thumb"])
		base["title"] = util.AsString(firstNonZero(content["description"], content["title"]))
	case "chat.sticker":
		base["extra"] = util.JSONString(map[string]any{"id": content["id"], "catId": content["catId"], "type": content["type"]})
	case "chat.recommended", "chat.link":
		extra := jsonMap(util.AsString(content["params"]))
		base["href"] = util.AsString(content["href"])
		base["thumb"] = util.AsString(content["thumb"])
		base["title"] = util.AsString(content["title"])
		base["linkCaption"] = "https://chat.zalo.me/"
		base["redirect_url"] = util.AsString(extra["redirect_url"])
		base["streamUrl"] = util.AsString(extra["streamUrl"])
		base["artist"] = util.AsString(extra["artist"])
		base["stream_icon"] = util.AsString(extra["stream_icon"])
		base["type"] = 2
		base["extra"] = util.JSONString(map[string]any{"action": content["action"], "params": util.JSONString(map[string]any{"mediaTitle": extra["mediaTitle"], "artist": extra["artist"], "src": extra["src"], "stream_icon": extra["stream_icon"], "streamUrl": extra["streamUrl"], "type": 2})})
	case "chat.location.new":
		base["title"] = util.AsString(firstNonZero(content["title"], content["description"]))
	case "share.file":
		extra := jsonMap(util.AsString(content["params"]))
		base["title"] = util.AsString(content["title"])
		base["extra"] = util.JSONString(map[string]any{"fileSize": "7295", "checksum": extra["checksum"], "fileExt": extra["fileExt"], "tWidth": extra["tWidth"], "tHeight": extra["tHeight"], "duration": extra["duration"], "fType": extra["fType"], "fdata": extra["fdata"]})
	case "chat.gif":
		base["thumb"] = util.AsString(content["thumb"])
	}
	payload["params"] = util.JSONString(base)
	form, err := m.EncodedForm(payload)
	if err != nil {
		return nil, err
	}
	data, err := m.PostJSON("https://groupboard-wpa.chat.zalo.me/api/board/topic/createv2", m.Query(map[string]any{"zpw_ver": 645, "zpw_type": m.APILoginType}), form)
	if err != nil {
		return nil, err
	}
	return m.parseGroup(data)
}

func (m *MessageAPI) UnpinMessage(pinID any, pinTime int64, groupID string) (*worker.Group, error) {
	enc, err := m.Encode(map[string]any{"grid": groupID, "imei": m.IMEIValue(), "topic": map[string]any{"topicId": util.AsString(pinID), "topicType": 2}, "boardVersion": pinTime})
	if err != nil {
		return nil, err
	}
	data, err := m.GetJSON("https://groupboard-wpa.chat.zalo.me/api/board/unpinv2", m.Query(map[string]any{"params": enc, "zpw_ver": 645, "zpw_type": m.APILoginType}))
	if err != nil {
		return nil, err
	}
	return m.parseGroup(data)
}

func firstNonZero(values ...any) any {
	for _, v := range values {
		if util.AsString(v) != "" && util.AsString(v) != "0" {
			return v
		}
	}
	return nil
}

var _ = json.Unmarshal

package socket

import (
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

func (s *SocketAPI) handleMessageObject(raw any, threadType core.ThreadType) {
	if s.UploadOnly() {
		return
	}
	msgMap := util.AsMap(raw)
	if len(msgMap) == 0 {
		msgMap = map[string]any{}
	}
	selfUID := s.accountUIDValue()
	uidFrom := normalizeSenderID(msgMap["uidFrom"], "")
	if uidFrom == "" {
		uidFrom = normalizeSenderID(msgMap["userId"], "")
	}
	if uidFrom == "" {
		uidFrom = normalizeSenderID(msgMap["uin"], "")
	}
	userThreadID := normalizeSenderID(firstNonEmpty(msgMap["idTo"], msgMap["toId"], msgMap["toid"]), "")
	if uidFrom == "" {
		uidFrom = selfUID
	}
	if uidFrom != "" {
		msgMap["uidFrom"] = uidFrom
		msgMap["userId"] = uidFrom
		msgMap["uin"] = uidFrom
	}
	msgObj := worker.MessageObjectFromDict(msgMap)
	threadID := uidFrom
	if threadType == core.GROUP {
		threadID = util.AsString(firstNonEmpty(msgObj.Get("idTo"), selfUID))
	} else {
		threadID = resolveDirectThreadID(uidFrom, normalizeSenderID(firstNonEmpty(msgObj.Get("idTo"), msgObj.Get("toId"), msgObj.Get("toid"), userThreadID), ""), selfUID)
	}
	message := util.AsString(msgObj.Get("content"))
	s.dispatch(func() {
		if s.Hub != nil {
			s.Hub.PublishMessage(worker.MessageEvent{MessageID: util.AsString(msgObj.Get("msgId")), UserID: uidFrom, ThreadID: threadID, ThreadType: threadType, Message: message, Data: msgObj})
		}
	})
}

func normalizeSenderID(value any, fallback string) string {
	text := strings.TrimSpace(util.AsString(value))
	if text == "" || text == "0" || text == "<nil>" {
		return strings.TrimSpace(fallback)
	}
	return text
}

func resolveDirectThreadID(uidFrom string, idTo string, selfUID string) string {
	uidFrom = strings.TrimSpace(uidFrom)
	idTo = strings.TrimSpace(idTo)
	selfUID = strings.TrimSpace(selfUID)

	switch {
	case uidFrom != "" && uidFrom != selfUID:
		return uidFrom
	case idTo != "" && idTo != selfUID:
		return idTo
	case uidFrom != "":
		return uidFrom
	case idTo != "":
		return idTo
	default:
		return selfUID
	}
}

func (s *SocketAPI) accountUIDValue() string {
	if s == nil {
		return ""
	}
	if s.State != nil {
		for _, key := range []string{"userId", "user_id", "uid"} {
			text := strings.TrimSpace(util.AsString(s.State.Config[key]))
			if text != "" && text != "0" && text != "<nil>" {
				return text
			}
		}
	}
	uid := strings.TrimSpace(s.UIDValue())
	if uid == "0" {
		return ""
	}
	return uid
}

func (s *SocketAPI) handleDecodedPacket(version byte, cmd int, subCmd byte, parsed map[string]any, parsedData map[string]any) bool {
	if version == 1 && cmd == 3000 && subCmd == 0 {
		s.publishError(fmt.Errorf("another connection is opened, closing this one"))
		_ = s.StopListening()
		return true
	}
	if version == 1 && cmd == 501 && subCmd == 0 {
		for _, item := range util.AsSlice(util.AsMap(parsedData["data"])["msgs"]) {
			s.handleMessageObject(item, core.USER)
		}
		return true
	}
	if version == 1 && cmd == 521 && subCmd == 0 {
		for _, item := range util.AsSlice(util.AsMap(parsedData["data"])["groupMsgs"]) {
			s.handleMessageObject(item, core.GROUP)
		}
		return true
	}
	if version == 1 && cmd == 601 && subCmd == 0 {
		for _, item := range util.AsSlice(util.AsMap(parsedData["data"])["controls"]) {
			content := util.AsMap(util.AsMap(item)["content"])
			actType := util.AsString(content["act_type"])
			if actType == "group" {
				if s.UploadOnly() {
					continue
				}
				act := util.AsString(content["act"])
				if act == "join_reject" {
					continue
				}
				raw := decodeStringJSON(content["data"])
				eventObj := worker.EventObjectFromDict(raw)
				eventType := worker.GroupEventType(util.GetGroupEventType(act))
				s.dispatch(func() {
					if s.Hub != nil {
						s.Hub.PublishGroupEvent(worker.GroupEventEnvelope{Event: eventObj, EventType: eventType})
					}
				})
				continue
			}
			if actType == "file_done" || actType == "voice_aac_success" {
				dataMap := util.AsMap(content["data"])
				fileURL := util.AsString(dataMap["url"])
				if fileURL == "" {
					fileURL = util.AsString(firstNonEmpty(dataMap["5"], dataMap["6"]))
				}
				if s.Hub != nil {
					s.Hub.PublishUpload(worker.UploadCompleteEvent{ActType: actType, FileID: util.AsString(content["fileId"]), FileURL: fileURL})
				}
				continue
			}
		}
		return true
	}
	if cmd == 612 {
		if s.UploadOnly() {
			return true
		}
		data612 := util.AsMap(parsedData["data"])
		for _, react := range util.AsSlice(data612["reacts"]) {
			m := util.AsMap(react)
			m["content"] = decodeStringJSON(m["content"])
			s.handleMessageObject(m, core.USER)
		}
		for _, react := range util.AsSlice(data612["reactGroups"]) {
			m := util.AsMap(react)
			m["content"] = decodeStringJSON(m["content"])
			s.handleMessageObject(m, core.GROUP)
		}
		return true
	}
	return false
}

func (s *SocketAPI) handleWsBinary(data []byte) {
	if len(data) < 5 {
		return
	}
	version, cmd, subCmd, err := util.GetHeader(data[:4])
	if err != nil {
		s.publishError(err)
		return
	}
	decodedData := string(data[4:])
	if decodedData == "" || strings.Contains(decodedData, "eventId") {
		return
	}
	parsed, err := util.DecodeJSONMap([]byte(decodedData))
	if err != nil {
		s.publishError(err)
		return
	}
	if version == 1 && cmd == 1 && subCmd == 1 && util.AsString(parsed["key"]) != "" {
		s.wsKey = util.AsString(parsed["key"])
		s.startPing()
		return
	}
	if s.wsKey == "" {
		s.publishError(fmt.Errorf("unable to decrypt data because key not found"))
		return
	}
	parsedData, err := util.ZWSDecode(parsed, s.wsKey)
	if err != nil {
		s.publishError(err)
		return
	}
	s.handleDecodedPacket(version, cmd, subCmd, parsed, parsedData)
}

func (s *SocketAPI) startPing() {
	if s.pingStopCh != nil {
		close(s.pingStopCh)
	}
	s.pingStopCh = make(chan struct{})
	go func(stopCh chan struct{}) {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.mu.Lock()
				conn := s.conn
				s.mu.Unlock()
				if conn != nil {
					_ = conn.WriteMessage(websocket.BinaryMessage, buildPingPayload())
				}
			case <-stopCh:
				return
			}
		}
	}(s.pingStopCh)
}

func (s *SocketAPI) Listen(thread bool, reconnect int) error {
	rawURLs, err := s.wsURLs()
	if err != nil {
		return err
	}
	s.thread = thread
	attempt := 0
	if reconnect < 0 {
		reconnect = 5
	}
	for {
		var conn *websocket.Conn
		var dialErr error
		for _, rawURL := range rawURLs {
			headers, err := s.wsHeaders(rawURL)
			if err != nil {
				dialErr = err
				continue
			}
			dialer := websocket.Dialer{HandshakeTimeout: 30 * time.Second, EnableCompression: false}
			conn, _, dialErr = dialer.Dial(rawURL, headers)
			if dialErr == nil {
				break
			}
		}
		if dialErr != nil {
			s.publishError(dialErr)
			attempt++
			if reconnect == 0 || attempt > reconnect {
				return dialErr
			}
			time.Sleep(time.Second)
			continue
		}
		s.mu.Lock()
		s.conn = conn
		s.wsKey = ""
		s.mu.Unlock()
		s.setListening(true)
		if s.Hub != nil && !s.UploadOnly() {
			s.Hub.PublishListening(worker.ListeningEvent{
				UserID:      s.UIDValue(),
				PhoneNumber: s.State.PhoneNumber,
				LoginType:   s.APILoginType,
				Timestamp:   util.Now(),
			})
		}
		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				s.mu.Lock()
				currentConn := s.conn
				currentListening := s.listening
				s.mu.Unlock()
				if currentListening && currentConn == conn {
					s.publishError(err)
				}
				break
			}
			if messageType == websocket.BinaryMessage {
				s.handleWsBinary(data)
			}
		}
		s.setListening(false)
		s.mu.Lock()
		if s.conn == conn {
			s.conn = nil
		}
		s.mu.Unlock()
		attempt++
		if reconnect == 0 || attempt > reconnect {
			return nil
		}
		time.Sleep(time.Second)
	}
}

func (s *SocketAPI) StopListening() error {
	s.mu.Lock()
	conn := s.conn
	s.conn = nil
	s.listening = false
	s.mu.Unlock()
	if s.pingStopCh != nil {
		close(s.pingStopCh)
		s.pingStopCh = nil
	}
	if s.Hub != nil {
		s.Hub.ClearWaiters()
	}
	if conn != nil {
		return conn.Close()
	}
	return nil
}

func firstNonEmpty(values ...any) any {
	for _, v := range values {
		if util.AsString(v) != "" && util.AsString(v) != "0" {
			return v
		}
	}
	return nil
}

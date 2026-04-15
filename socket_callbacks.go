package zago

import (
	"log"
	"strings"

	core "github.com/tranhaonguyendev/za-go/internal/core"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

type SocketCallbacks struct {
	Message   func(string, string, string, *worker.MessageObject, string, core.ThreadType)
	Event     func(*worker.EventObject, worker.GroupEventType)
	Delivered func(any, string, core.ThreadType, int64)
	Seen      func(any, string, core.ThreadType, int64)
	Error     func(error, int64)
	Upload    func(worker.UploadCompleteEvent)
	Listening func(worker.ListeningEvent)
}

func (z *ZaloAPI) bindSocketCallbacks() {
	if z == nil || z.hub == nil {
		return
	}
	z.hub.SetMessageHandler(func(evt worker.MessageEvent) {
		handler := z.getMessageListener()
		if handler == nil {
			z.MessageListener(evt.MessageID, evt.UserID, evt.Message, evt.Data, evt.ThreadID, evt.ThreadType)
			return
		}
		handler(evt.MessageID, evt.UserID, evt.Message, evt.Data, evt.ThreadID, evt.ThreadType)
	})
	z.hub.SetGroupHandler(func(evt worker.GroupEventEnvelope) {
		handler := z.getEventListener()
		if handler == nil {
			z.EventListener(evt.Event, evt.EventType)
			return
		}
		handler(evt.Event, evt.EventType)
	})
	z.hub.SetDeliveryHandler(func(evt worker.DeliveryEvent) {
		handler := z.getDeliveredListener()
		if handler == nil {
			z.MessageListenerDelivered(evt.MsgIDs, evt.ThreadID, evt.ThreadType, evt.Timestamp)
			return
		}
		handler(evt.MsgIDs, evt.ThreadID, evt.ThreadType, evt.Timestamp)
	})
	z.hub.SetSeenHandler(func(evt worker.SeenEvent) {
		handler := z.getSeenListener()
		if handler == nil {
			z.OnMarkedSeen(evt.MsgIDs, evt.ThreadID, evt.ThreadType, evt.Timestamp)
			return
		}
		handler(evt.MsgIDs, evt.ThreadID, evt.ThreadType, evt.Timestamp)
	})
	z.hub.SetErrorHandler(func(evt worker.SocketErrorEvent) {
		handler := z.getErrorListener()
		if handler == nil {
			z.OnErrorCallBack(evt.Err, evt.Timestamp)
			return
		}
		handler(evt.Err, evt.Timestamp)
	})
	z.hub.SetUploadHandler(func(evt worker.UploadCompleteEvent) {
		handler := z.getUploadListener()
		if handler == nil {
			z.UploadListener(evt)
			return
		}
		handler(evt)
	})
	z.hub.SetListeningHandler(func(evt worker.ListeningEvent) {
		handler := z.getListeningListener()
		if handler == nil {
			z.OnListening(evt)
			return
		}
		handler(evt)
	})
}

func (z *ZaloAPI) SetSocketCallbacks(callbacks SocketCallbacks) {
	z.SetMessageListener(callbacks.Message)
	z.SetEventListener(callbacks.Event)
	z.SetDeliveredListener(callbacks.Delivered)
	z.SetSeenListener(callbacks.Seen)
	z.SetErrorListener(callbacks.Error)
	z.SetUploadListener(callbacks.Upload)
	z.SetListeningListener(callbacks.Listening)
}

func (z *ZaloAPI) SetMessageListener(fn func(string, string, string, *worker.MessageObject, string, core.ThreadType)) {
	z.listenerMu.Lock()
	defer z.listenerMu.Unlock()
	z.onMessage = fn
}

func (z *ZaloAPI) SetEventListener(fn func(*worker.EventObject, worker.GroupEventType)) {
	z.listenerMu.Lock()
	defer z.listenerMu.Unlock()
	z.onEvent = fn
}

func (z *ZaloAPI) SetDeliveredListener(fn func(any, string, core.ThreadType, int64)) {
	z.listenerMu.Lock()
	defer z.listenerMu.Unlock()
	z.onDelivered = fn
}

func (z *ZaloAPI) SetSeenListener(fn func(any, string, core.ThreadType, int64)) {
	z.listenerMu.Lock()
	defer z.listenerMu.Unlock()
	z.onSeen = fn
}

func (z *ZaloAPI) SetErrorListener(fn func(error, int64)) {
	z.listenerMu.Lock()
	defer z.listenerMu.Unlock()
	z.onError = fn
}

func (z *ZaloAPI) SetErrorCallback(fn func(error, int64)) {
	z.SetErrorListener(fn)
}

func (z *ZaloAPI) SetUploadListener(fn func(worker.UploadCompleteEvent)) {
	z.listenerMu.Lock()
	defer z.listenerMu.Unlock()
	z.onUpload = fn
}

func (z *ZaloAPI) SetListeningListener(fn func(worker.ListeningEvent)) {
	z.listenerMu.Lock()
	defer z.listenerMu.Unlock()
	z.onListening = fn
}

func (z *ZaloAPI) getMessageListener() func(string, string, string, *worker.MessageObject, string, core.ThreadType) {
	z.listenerMu.RLock()
	defer z.listenerMu.RUnlock()
	return z.onMessage
}

func (z *ZaloAPI) getEventListener() func(*worker.EventObject, worker.GroupEventType) {
	z.listenerMu.RLock()
	defer z.listenerMu.RUnlock()
	return z.onEvent
}

func (z *ZaloAPI) getDeliveredListener() func(any, string, core.ThreadType, int64) {
	z.listenerMu.RLock()
	defer z.listenerMu.RUnlock()
	return z.onDelivered
}

func (z *ZaloAPI) getSeenListener() func(any, string, core.ThreadType, int64) {
	z.listenerMu.RLock()
	defer z.listenerMu.RUnlock()
	return z.onSeen
}

func (z *ZaloAPI) getErrorListener() func(error, int64) {
	z.listenerMu.RLock()
	defer z.listenerMu.RUnlock()
	return z.onError
}

func (z *ZaloAPI) getUploadListener() func(worker.UploadCompleteEvent) {
	z.listenerMu.RLock()
	defer z.listenerMu.RUnlock()
	return z.onUpload
}

func (z *ZaloAPI) getListeningListener() func(worker.ListeningEvent) {
	z.listenerMu.RLock()
	defer z.listenerMu.RUnlock()
	return z.onListening
}

func (z *ZaloAPI) MessageListener(mid string, userID string, message string, data *worker.MessageObject, threadID string, threadType core.ThreadType) {
	log.Printf("%s from %s in %s", message, threadID, threadType.String())
}

func (z *ZaloAPI) EventListener(eventData *worker.EventObject, eventType worker.GroupEventType) {
}

func (z *ZaloAPI) MessageListenerDelivered(msgIDs any, threadID string, threadType core.ThreadType, ts int64) {
	log.Printf("Marked messages %v as delivered in [(%s, %s)] at %d.", msgIDs, threadID, threadType.String(), normalizeListenerTimestamp(ts))
}

func (z *ZaloAPI) OnMarkedSeen(msgIDs any, threadID string, threadType core.ThreadType, ts int64) {
	log.Printf("Marked messages %v as seen in [(%s, %s)] at %d.", msgIDs, threadID, threadType.String(), normalizeListenerTimestamp(ts))
}

func (z *ZaloAPI) UploadListener(evt worker.UploadCompleteEvent) {
}

func (z *ZaloAPI) OnListening(evt worker.ListeningEvent) {
	log.Printf("Listening websocket for %s [%d] [%s].", evt.UserID, evt.LoginType, evt.PhoneNumber)
}

func (z *ZaloAPI) OnErrorCallBack(err error, ts int64) {
	if err == nil {
		return
	}
	if strings.Contains(strings.ToLower(err.Error()), "login") {
		log.Printf("[ERROR #%d] Imei Cookies Has Expired Or Null Error To Run!", ts)
		return
	}
	log.Printf("An error occurred at %d: %v", ts, err)
}

func normalizeListenerTimestamp(ts int64) int64 {
	if ts <= 0 {
		return 0
	}
	return ts / 1000
}

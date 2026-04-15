package worker

import (
	"sync"
	"time"

	core "github.com/tranhaonguyendev/za-go/internal/core"
)

type GroupEventType string
type EventType string

const (
	GroupEventJoin          GroupEventType = "join"
	GroupEventLeave         GroupEventType = "leave"
	GroupEventUpdate        GroupEventType = "update"
	GroupEventUnknown       GroupEventType = "unknown"
	GroupEventReaction      GroupEventType = "reaction"
	GroupEventNewLink       GroupEventType = "new_link"
	GroupEventAddAdmin      GroupEventType = "add_admin"
	GroupEventRemoveAdmin   GroupEventType = "remove_admin"
	GroupEventJoinRequest   GroupEventType = "join_request"
	GroupEventBlockMember   GroupEventType = "block_member"
	GroupEventRemoveMember  GroupEventType = "remove_member"
	GroupEventUpdateSetting GroupEventType = "update_setting"
)

const (
	EventReaction EventType = "reaction"
)

type MessageEvent struct {
	MessageID  string
	UserID     string
	ThreadID   string
	ThreadType core.ThreadType
	Message    string
	Data       *MessageObject
}

type GroupEventEnvelope struct {
	Event     *EventObject
	EventType GroupEventType
}

type DeliveryEvent struct {
	MsgIDs     any
	ThreadID   string
	ThreadType core.ThreadType
	Timestamp  int64
}

type SeenEvent struct {
	MsgIDs     any
	ThreadID   string
	ThreadType core.ThreadType
	Timestamp  int64
}

type SocketErrorEvent struct {
	Err       error
	Timestamp int64
}

type UploadCompleteEvent struct {
	ActType string
	FileID  string
	FileURL string
}

type ListeningEvent struct {
	UserID      string
	PhoneNumber string
	LoginType   int
	Timestamp   int64
}

type cachedUpload struct {
	Event      UploadCompleteEvent
	ReceivedAt time.Time
}

type Hub struct {
	messageCh  chan MessageEvent
	groupCh    chan GroupEventEnvelope
	deliveryCh chan DeliveryEvent
	seenCh     chan SeenEvent
	errorCh    chan SocketErrorEvent
	uploadCh   chan UploadCompleteEvent
	listenCh   chan ListeningEvent

	mu              sync.Mutex
	callbacksMu     sync.RWMutex
	waiters         map[string]chan UploadCompleteEvent
	uploadCache     map[string]cachedUpload
	messageHandler  func(MessageEvent)
	groupHandler    func(GroupEventEnvelope)
	deliveryHandler func(DeliveryEvent)
	seenHandler     func(SeenEvent)
	errorHandler    func(SocketErrorEvent)
	uploadHandler   func(UploadCompleteEvent)
	listenHandler   func(ListeningEvent)
}

func NewHub() *Hub {
	return &Hub{
		messageCh:   make(chan MessageEvent, 256),
		groupCh:     make(chan GroupEventEnvelope, 128),
		deliveryCh:  make(chan DeliveryEvent, 128),
		seenCh:      make(chan SeenEvent, 128),
		errorCh:     make(chan SocketErrorEvent, 128),
		uploadCh:    make(chan UploadCompleteEvent, 128),
		listenCh:    make(chan ListeningEvent, 64),
		waiters:     map[string]chan UploadCompleteEvent{},
		uploadCache: map[string]cachedUpload{},
	}
}

func (h *Hub) MessageEvents() <-chan MessageEvent       { return h.messageCh }
func (h *Hub) GroupEvents() <-chan GroupEventEnvelope   { return h.groupCh }
func (h *Hub) DeliveryEvents() <-chan DeliveryEvent     { return h.deliveryCh }
func (h *Hub) SeenEvents() <-chan SeenEvent             { return h.seenCh }
func (h *Hub) SocketErrors() <-chan SocketErrorEvent    { return h.errorCh }
func (h *Hub) UploadEvents() <-chan UploadCompleteEvent { return h.uploadCh }
func (h *Hub) ListeningEvents() <-chan ListeningEvent   { return h.listenCh }

func safeCall(fn func()) {
	defer func() { _ = recover() }()
	fn()
}

func (h *Hub) SetMessageHandler(fn func(MessageEvent)) {
	h.callbacksMu.Lock()
	defer h.callbacksMu.Unlock()
	h.messageHandler = fn
}

func (h *Hub) SetGroupHandler(fn func(GroupEventEnvelope)) {
	h.callbacksMu.Lock()
	defer h.callbacksMu.Unlock()
	h.groupHandler = fn
}

func (h *Hub) SetDeliveryHandler(fn func(DeliveryEvent)) {
	h.callbacksMu.Lock()
	defer h.callbacksMu.Unlock()
	h.deliveryHandler = fn
}

func (h *Hub) SetSeenHandler(fn func(SeenEvent)) {
	h.callbacksMu.Lock()
	defer h.callbacksMu.Unlock()
	h.seenHandler = fn
}

func (h *Hub) SetErrorHandler(fn func(SocketErrorEvent)) {
	h.callbacksMu.Lock()
	defer h.callbacksMu.Unlock()
	h.errorHandler = fn
}

func (h *Hub) SetUploadHandler(fn func(UploadCompleteEvent)) {
	h.callbacksMu.Lock()
	defer h.callbacksMu.Unlock()
	h.uploadHandler = fn
}

func (h *Hub) SetListeningHandler(fn func(ListeningEvent)) {
	h.callbacksMu.Lock()
	defer h.callbacksMu.Unlock()
	h.listenHandler = fn
}

func (h *Hub) PublishMessage(evt MessageEvent) {
	select {
	case h.messageCh <- evt:
	default:
	}
	h.callbacksMu.RLock()
	handler := h.messageHandler
	h.callbacksMu.RUnlock()
	if handler != nil {
		safeCall(func() { handler(evt) })
	}
}

func (h *Hub) PublishGroupEvent(evt GroupEventEnvelope) {
	select {
	case h.groupCh <- evt:
	default:
	}
	h.callbacksMu.RLock()
	handler := h.groupHandler
	h.callbacksMu.RUnlock()
	if handler != nil {
		safeCall(func() { handler(evt) })
	}
}

func (h *Hub) PublishDelivery(evt DeliveryEvent) {
	select {
	case h.deliveryCh <- evt:
	default:
	}
	h.callbacksMu.RLock()
	handler := h.deliveryHandler
	h.callbacksMu.RUnlock()
	if handler != nil {
		safeCall(func() { handler(evt) })
	}
}

func (h *Hub) PublishSeen(evt SeenEvent) {
	select {
	case h.seenCh <- evt:
	default:
	}
	h.callbacksMu.RLock()
	handler := h.seenHandler
	h.callbacksMu.RUnlock()
	if handler != nil {
		safeCall(func() { handler(evt) })
	}
}

func (h *Hub) PublishError(evt SocketErrorEvent) {
	select {
	case h.errorCh <- evt:
	default:
	}
	h.callbacksMu.RLock()
	handler := h.errorHandler
	h.callbacksMu.RUnlock()
	if handler != nil {
		safeCall(func() { handler(evt) })
	}
}

func (h *Hub) PublishUpload(evt UploadCompleteEvent) {
	select {
	case h.uploadCh <- evt:
	default:
	}
	h.mu.Lock()
	h.pruneUploadCacheLocked(time.Now())
	if evt.FileID != "" {
		h.uploadCache[evt.FileID] = cachedUpload{Event: evt, ReceivedAt: time.Now()}
	}
	if ch, ok := h.waiters[evt.FileID]; ok {
		select {
		case ch <- evt:
		default:
		}
		close(ch)
		delete(h.waiters, evt.FileID)
		delete(h.uploadCache, evt.FileID)
	}
	h.mu.Unlock()
	h.callbacksMu.RLock()
	handler := h.uploadHandler
	h.callbacksMu.RUnlock()
	if handler != nil {
		safeCall(func() { handler(evt) })
	}
}

func (h *Hub) RegisterUploadWaiter(fileID string) <-chan UploadCompleteEvent {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch := make(chan UploadCompleteEvent, 1)
	h.pruneUploadCacheLocked(time.Now())
	if cached, ok := h.uploadCache[fileID]; ok {
		ch <- cached.Event
		close(ch)
		delete(h.uploadCache, fileID)
		return ch
	}
	h.waiters[fileID] = ch
	return ch
}

func (h *Hub) PublishListening(evt ListeningEvent) {
	select {
	case h.listenCh <- evt:
	default:
	}
	h.callbacksMu.RLock()
	handler := h.listenHandler
	h.callbacksMu.RUnlock()
	if handler != nil {
		safeCall(func() { handler(evt) })
	}
}

func (h *Hub) CancelUploadWaiter(fileID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if ch, ok := h.waiters[fileID]; ok {
		close(ch)
		delete(h.waiters, fileID)
	}
}

func (h *Hub) ClearWaiters() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for id, ch := range h.waiters {
		close(ch)
		delete(h.waiters, id)
	}
}

func (h *Hub) pruneUploadCacheLocked(now time.Time) {
	const ttl = 30 * time.Second
	for fileID, cached := range h.uploadCache {
		if now.Sub(cached.ReceivedAt) > ttl {
			delete(h.uploadCache, fileID)
		}
	}
}

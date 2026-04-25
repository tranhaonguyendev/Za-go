package zago

import (
	socketapi "github.com/tranhaonguyendev/za-go/internal/api/socket"
	core "github.com/tranhaonguyendev/za-go/internal/core"
	"github.com/tranhaonguyendev/za-go/internal/worker"
)

type Message = worker.Message
type ThreadType = core.ThreadType
type User = worker.User
type Group = worker.Group
type MessageObject = worker.MessageObject
type ContextObject = worker.ContextObject
type EventObject = worker.EventObject
type GroupEventType = worker.GroupEventType
type EventType = worker.EventType
type MessageEvent = worker.MessageEvent
type GroupEventEnvelope = worker.GroupEventEnvelope
type DeliveryEvent = worker.DeliveryEvent
type SeenEvent = worker.SeenEvent
type SocketErrorEvent = worker.SocketErrorEvent
type UploadCompleteEvent = worker.UploadCompleteEvent
type ListeningEvent = worker.ListeningEvent
type QRAuthResult = socketapi.QRAuthResult

var MessageStyle = worker.MessageStyle
var MultiMsgStyle = worker.MultiMsgStyle
var MessageReaction = worker.MessageReaction
var Mention = worker.Mention
var NewUser = worker.UserFromDict
var NewGroup = worker.GroupFromDict
var NewContextObject = worker.ContextFromDict
var NewMessageObject = worker.MessageObjectFromDict
var NewEventObject = worker.EventObjectFromDict

const (
	ThreadTypeUSER  = core.USER
	ThreadTypeGROUP = core.GROUP
)

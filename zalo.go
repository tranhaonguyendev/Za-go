package zago

import (
	"fmt"
	"strings"
	"sync"
	"time"

	getapi "github.com/nguyendev/zago/internal/api/get"
	groupapi "github.com/nguyendev/zago/internal/api/group"
	groupaction "github.com/nguyendev/zago/internal/api/group/action"
	groupmessage "github.com/nguyendev/zago/internal/api/group/message"
	groupstatus "github.com/nguyendev/zago/internal/api/group/status"
	handle "github.com/nguyendev/zago/internal/api/handle"
	properties "github.com/nguyendev/zago/internal/api/properties"
	socketapi "github.com/nguyendev/zago/internal/api/socket"
	"github.com/nguyendev/zago/internal/app"
	authcore "github.com/nguyendev/zago/internal/auth"
	core "github.com/nguyendev/zago/internal/core"
	"github.com/nguyendev/zago/internal/worker"
)

const LoginAPI = 24

type ZaloAPI struct {
	state        *app.State
	hub          *worker.Hub
	auth         *authcore.LoginAuth
	send         *handle.SendAPI
	get          *getapi.GetAPI
	properties   *properties.PropertiesAPI
	group        *groupapi.GroupAPI
	groupAction  *groupaction.ActionAPI
	groupMessage *groupmessage.MessageAPI
	groupStatus  *groupstatus.StatusAPI
	socket       *socketapi.SocketAPI
	apiLoginType int
	imei         string
	uid          string
	refreshMu    sync.Mutex
	syncedIMEI   string
	syncedUID    string
	accountMu    sync.RWMutex
	accountID    string
	accountName  string
	listenerMu   sync.RWMutex
	onMessage    func(string, string, string, *worker.MessageObject, string, core.ThreadType)
	onEvent      func(*worker.EventObject, worker.GroupEventType)
	onDelivered  func(any, string, core.ThreadType, int64)
	onSeen       func(any, string, core.ThreadType, int64)
	onError      func(error, int64)
	onUpload     func(worker.UploadCompleteEvent)
	onListening  func(worker.ListeningEvent)
}

func Zalo(phone, password, imei string, sessionCookies any, userAgent string, autoLogin bool, login int) (*ZaloAPI, error) {
	if login == 0 {
		login = LoginAPI
	}
	state := app.NewState()
	hub := worker.NewHub()
	getSvc := getapi.NewGetAPI(state, login, hub)
	z := &ZaloAPI{
		state:        state,
		hub:          hub,
		apiLoginType: login,
		auth:         authcore.NewLoginAuth(state),
		get:          getSvc,
		send:         handle.NewSendAPI(state, login, hub),
		properties:   properties.NewPropertiesAPI(state, login, hub),
		group:        groupapi.NewGroupAPI(state, login, hub, getSvc),
		groupAction:  groupaction.NewActionAPI(state, login, hub),
		groupMessage: groupmessage.NewMessageAPI(state, login, hub),
		groupStatus:  groupstatus.NewStatusAPI(state, login, hub),
		socket:       socketapi.NewSocketAPI(state, login, hub),
	}
	if sessionCookies != nil {
		z.SetSession(sessionCookies)
	}
	if autoLogin && !z.IsLoggedIn() {
		if err := z.Login(phone, password, imei, userAgent); err != nil {
			return nil, err
		}
	}
	z.bindSocketCallbacks()
	z.refreshServices()
	return z, nil
}

func (z *ZaloAPI) refreshServices() {
	imei := strings.TrimSpace(z.imei)
	if imei == "" {
		imei = strings.TrimSpace(z.state.ClientUUID)
		z.imei = imei
	}
	uid := strings.TrimSpace(z.uid)
	if uid == "" {
		uid = strings.TrimSpace(z.state.UserClientID)
		z.uid = uid
	}

	z.refreshMu.Lock()
	defer z.refreshMu.Unlock()
	if imei == z.syncedIMEI && uid == z.syncedUID {
		return
	}
	z.syncedIMEI, z.syncedUID = imei, uid

	z.send.IMEI, z.send.UID = imei, uid
	z.send.Uploader.IMEI, z.send.Uploader.UID = imei, uid
	z.get.IMEI, z.get.UID = imei, uid
	z.properties.IMEI, z.properties.UID = imei, uid
	z.group.IMEI, z.group.UID = imei, uid
	z.groupAction.IMEI, z.groupAction.UID = imei, uid
	z.groupMessage.IMEI, z.groupMessage.UID = imei, uid
	z.groupStatus.IMEI, z.groupStatus.UID = imei, uid
	z.socket.IMEI, z.socket.UID = imei, uid
}

func (z *ZaloAPI) IsLoggedIn() bool { return z.auth.IsLoggedIn() }

func (z *ZaloAPI) cacheAccountProfile(raw any) {
	obj, ok := raw.(interface{ ToMap() map[string]any })
	if !ok {
		return
	}
	data := obj.ToMap()
	profile := data
	if nested, ok := data["profile"].(map[string]any); ok {
		profile = nested
	}
	userID := strings.TrimSpace(fmt.Sprintf("%v", profile["userId"]))
	if userID == "<nil>" {
		userID = ""
	}
	name := ""
	for _, key := range []string{"zaloName", "displayName", "name", "username", "phoneNumber"} {
		if value := strings.TrimSpace(fmt.Sprintf("%v", profile[key])); value != "" && value != "<nil>" {
			name = value
			break
		}
	}
	z.accountMu.Lock()
	if userID != "" {
		z.accountID = userID
		if z.state != nil {
			z.state.Config["userId"] = userID
			z.state.Config["user_id"] = userID
			z.state.Config["uid"] = userID
		}
	}
	if name != "" {
		z.accountName = name
	}
	z.accountMu.Unlock()
}

func (z *ZaloAPI) ensureAccountProfile() {
	z.accountMu.RLock()
	ready := z.accountID != "" || z.accountName != ""
	z.accountMu.RUnlock()
	if ready {
		return
	}
	raw, err := z.FetchAccountInfo()
	if err != nil {
		return
	}
	z.cacheAccountProfile(raw)
}

func (z *ZaloAPI) ClientID() string {
	z.refreshServices()
	return z.uid
}

func (z *ZaloAPI) UID() string {
	return z.UserID()
}

func (z *ZaloAPI) UserID() string {
	z.ensureAccountProfile()
	z.accountMu.RLock()
	defer z.accountMu.RUnlock()
	if z.accountID != "" {
		return z.accountID
	}
	z.refreshServices()
	return z.uid
}

func (z *ZaloAPI) AccountName() string {
	z.ensureAccountProfile()
	z.accountMu.RLock()
	defer z.accountMu.RUnlock()
	return z.accountName
}

func (z *ZaloAPI) SetSession(sessionCookies any) bool {
	ok := z.auth.SetSession(sessionCookies)
	if ok && z.state.UserClientID != "" {
		z.uid = z.state.UserClientID
	}
	z.accountMu.Lock()
	z.accountID = ""
	z.accountName = ""
	z.accountMu.Unlock()
	z.refreshServices()
	return ok
}

func (z *ZaloAPI) Login(phone, password, imei, userAgent string) error {
	if strings.TrimSpace(imei) == "" {
		return fmt.Errorf("imei is required")
	}
	if err := z.state.Login(phone, password, imei, nil, userAgent); err != nil {
		return err
	}
	z.imei = z.state.ClientUUID
	if z.imei == "" {
		z.imei = imei
	}
	z.uid = z.state.UserClientID
	z.accountMu.Lock()
	z.accountID = ""
	z.accountName = ""
	z.accountMu.Unlock()
	z.refreshServices()
	return nil
}

func (z *ZaloAPI) MessageEvents() <-chan worker.MessageEvent       { return z.hub.MessageEvents() }
func (z *ZaloAPI) GroupEvents() <-chan worker.GroupEventEnvelope   { return z.hub.GroupEvents() }
func (z *ZaloAPI) DeliveryEvents() <-chan worker.DeliveryEvent     { return z.hub.DeliveryEvents() }
func (z *ZaloAPI) SeenEvents() <-chan worker.SeenEvent             { return z.hub.SeenEvents() }
func (z *ZaloAPI) SocketErrors() <-chan worker.SocketErrorEvent    { return z.hub.SocketErrors() }
func (z *ZaloAPI) UploadEvents() <-chan worker.UploadCompleteEvent { return z.hub.UploadEvents() }
func (z *ZaloAPI) ListeningEvents() <-chan worker.ListeningEvent   { return z.hub.ListeningEvents() }

func (z *ZaloAPI) Listen(thread bool, reconnect int) error {
	z.refreshServices()
	return z.socket.Listen(thread, reconnect)
}
func (z *ZaloAPI) StopListening() error                         { return z.socket.StopListening() }
func (z *ZaloAPI) AuthQRCode() (*socketapi.QRAuthResult, error) { return z.socket.AuthQRCode() }
func (z *ZaloAPI) WaitQRCodeScan(qr *socketapi.QRAuthResult, maxAttempts int, interval time.Duration) (bool, error) {
	return z.socket.WaitQRCodeScan(qr, maxAttempts, interval)
}
func (z *ZaloAPI) CheckQRCodeScan(qr *socketapi.QRAuthResult) (map[string]any, error) {
	return z.socket.CheckQRCodeScan(qr)
}
func (z *ZaloAPI) WaitQRCodeConfirm(qr *socketapi.QRAuthResult, maxAttempts int, interval time.Duration) (map[string]string, error) {
	return z.socket.WaitQRCodeConfirm(qr, maxAttempts, interval)
}
func (z *ZaloAPI) CheckQRCodeConfirm(qr *socketapi.QRAuthResult) (map[string]any, error) {
	return z.socket.CheckQRCodeConfirm(qr)
}
func (z *ZaloAPI) CheckQRSession(qr *socketapi.QRAuthResult) (string, error) {
	return z.socket.CheckQRSession(qr)
}
func (z *ZaloAPI) FetchQRUserInfo(qr *socketapi.QRAuthResult) (map[string]any, error) {
	return z.socket.FetchQRUserInfo(qr)
}
func (z *ZaloAPI) IsListening() bool { return z.socket.Listening() }

func (z *ZaloAPI) SendMessage(message worker.Message, threadID string, threadType core.ThreadType) (any, error) {
	z.refreshServices()
	return z.send.SendMessage(message, threadID, threadType)
}
func (z *ZaloAPI) SendVoice(voiceURL string, threadID string, threadType core.ThreadType, fileSize int, ttl int) (any, error) {
	z.refreshServices()
	return z.send.SendVoice(voiceURL, threadID, threadType, fileSize, ttl)
}
func (z *ZaloAPI) SendVideo(videoURL string, thumbnailURL string, duration int, threadID string, threadType core.ThreadType, width int, height int, message *worker.Message, ttl int) (any, error) {
	z.refreshServices()
	return z.send.SendVideo(videoURL, thumbnailURL, duration, threadID, threadType, width, height, message, ttl)
}
func (z *ZaloAPI) SendReaction(messageObject *worker.MessageObject, reactionIcon string, threadID string, threadType core.ThreadType, reactionType int) (any, error) {
	z.refreshServices()
	return z.send.SendReaction(messageObject, reactionIcon, threadID, threadType, reactionType)
}
func (z *ZaloAPI) SendImage(imageURL string, threadID string, threadType core.ThreadType, width int, height int, message *worker.Message, ttl int) (any, error) {
	z.refreshServices()
	return z.send.SendImage(imageURL, threadID, threadType, width, height, message, ttl)
}
func (z *ZaloAPI) SendFile(fileURL string, threadID string, threadType core.ThreadType, fileName string, fileSize int, extension string, ttl int, localPath string) (any, error) {
	z.refreshServices()
	return z.send.SendFile(fileURL, threadID, threadType, fileName, fileSize, extension, ttl, localPath)
}
func (z *ZaloAPI) AddFriend(userID string, msg string, language string) (*worker.User, error) {
	z.refreshServices()
	return z.send.AddFriend(userID, msg, language)
}
func (z *ZaloAPI) BlockUser(userID string) (*worker.User, error) {
	z.refreshServices()
	return z.send.BlockUser(userID)
}
func (z *ZaloAPI) UnblockUser(userID string) (*worker.User, error) {
	z.refreshServices()
	return z.send.UnblockUser(userID)
}
func (z *ZaloAPI) BlockViewFeed(userID string, isBlockFeed bool) (*worker.User, error) {
	z.refreshServices()
	return z.send.BlockViewFeed(userID, isBlockFeed)
}
func (z *ZaloAPI) UnfriendUser(userID string, language string) (any, error) {
	z.refreshServices()
	return z.send.UnfriendUser(userID, language)
}
func (z *ZaloAPI) SetAlias(friendID string, alias string) (any, error) {
	z.refreshServices()
	return z.send.SetAlias(friendID, alias)
}
func (z *ZaloAPI) SendBusinessCard(userID string, qrCodeURL string, threadID string, threadType core.ThreadType, phone string, ttl int) (any, error) {
	z.refreshServices()
	return z.send.SendBusinessCard(userID, qrCodeURL, threadID, threadType, phone, ttl)
}
func (z *ZaloAPI) SendReport(userID string, threadType core.ThreadType, reason int, content string) (any, error) {
	z.refreshServices()
	return z.send.SendReport(userID, threadType, reason, content)
}
func (z *ZaloAPI) SendSticker(stickerType int, stickerID int, cateID int, threadID string, threadType core.ThreadType, ttl int) (any, error) {
	z.refreshServices()
	return z.send.SendSticker(stickerType, stickerID, cateID, threadID, threadType, ttl)
}
func (z *ZaloAPI) SendMultiReaction(messageObject *worker.MessageObject, reactionIcon string, threadID string, threadType core.ThreadType, reactionType int, numreact int) (any, error) {
	z.refreshServices()
	return z.send.SendMultiReaction(messageObject, reactionIcon, threadID, threadType, reactionType, numreact)
}
func (z *ZaloAPI) SendCall(targetID string, callID string, extras ...string) (map[string]any, error) {
	z.refreshServices()
	return z.send.SendCall(targetID, callID, extras...)
}
func (z *ZaloAPI) CallGroupRequest(groupID string, userIDs []string, callID any, groupNames ...string) (map[string]any, error) {
	z.refreshServices()
	return z.send.CallGroupRequest(groupID, userIDs, callID, groupNames...)
}
func (z *ZaloAPI) CallGroupAdd(userIDs []string, callID any, hostCall any, groupID string) (map[string]any, error) {
	z.refreshServices()
	return z.send.CallGroupAdd(userIDs, callID, hostCall, groupID)
}
func (z *ZaloAPI) CallGroupCancel(callID any, hostCall any, groupID string) (map[string]any, error) {
	z.refreshServices()
	return z.send.CallGroupCancel(callID, hostCall, groupID)
}
func (z *ZaloAPI) CallGroup(groupID string, userIDs []string, extras ...any) (map[string]any, error) {
	z.refreshServices()
	return z.send.CallGroup(groupID, userIDs, extras...)
}
func (z *ZaloAPI) SendCustomSticker(staticImgURL string, animationImgURL string, threadID string, threadType core.ThreadType, reply string, width int, height int, ttl int, ai bool) (any, error) {
	z.refreshServices()
	return z.send.SendCustomSticker(staticImgURL, animationImgURL, threadID, threadType, reply, width, height, ttl, ai)
}
func (z *ZaloAPI) SendLink(linkURL string, threadID string, threadType core.ThreadType, message *worker.Message, ttl int) (any, error) {
	z.refreshServices()
	return z.send.SendLink(linkURL, threadID, threadType, message, ttl)
}
func (z *ZaloAPI) SendLocalGif(gifPath string, thumbnailURL string, threadID string, threadType core.ThreadType, gifName string, width int, height int, ttl int) (any, error) {
	z.refreshServices()
	return z.send.SendLocalGif(gifPath, thumbnailURL, threadID, threadType, gifName, width, height, ttl)
}
func (z *ZaloAPI) SendGifphy(gifPath string, thumbnailURL string, threadID string, threadType core.ThreadType, gifName string, width int, height int, ttl int) (any, error) {
	z.refreshServices()
	return z.send.SendGifphy(gifPath, thumbnailURL, threadID, threadType, gifName, width, height, ttl)
}
func (z *ZaloAPI) SendLocalImage(imagePath string, threadID string, threadType core.ThreadType, width int, height int, message *worker.Message, customPayload map[string]any, ttl int) (any, error) {
	z.refreshServices()
	return z.send.SendLocalImage(imagePath, threadID, threadType, width, height, message, customPayload, ttl)
}
func (z *ZaloAPI) SendMultiImage(imageURLs []string, threadID string, threadType core.ThreadType, width int, height int, message *worker.Message, ttl int) ([]any, error) {
	z.refreshServices()
	return z.send.SendMultiImage(imageURLs, threadID, threadType, width, height, message, ttl)
}
func (z *ZaloAPI) SendMultiLocalImage(imagePaths []string, threadID string, threadType core.ThreadType, width int, height int, message *worker.Message, ttl int) ([]any, error) {
	z.refreshServices()
	return z.send.SendMultiLocalImage(imagePaths, threadID, threadType, width, height, message, ttl)
}
func (z *ZaloAPI) SendCardBank(msgData *worker.MessageObject, bankNum string, nameAccBank string, bank string, threadID string, threadType core.ThreadType) (any, error) {
	z.refreshServices()
	return z.send.SendCardBank(msgData, bankNum, nameAccBank, bank, threadID, threadType)
}

func (z *ZaloAPI) UploadImage(filePath string, threadID string, threadType core.ThreadType) (map[string]any, error) {
	z.refreshServices()
	return z.send.Uploader.UploadImage(filePath, threadID, threadType)
}
func (z *ZaloAPI) UploadAttachment(filePath string, threadID string, threadType core.ThreadType) (map[string]any, error) {
	z.refreshServices()
	startedListener := false
	if !z.socket.Listening() {
		startedListener = true
		z.socket.SetUploadOnly(true)
		go func() { _ = z.socket.Listen(false, 0) }()
		deadline := time.Now().Add(5 * time.Second)
		for !z.socket.Ready() && time.Now().Before(deadline) {
			time.Sleep(100 * time.Millisecond)
		}
	}
	result, err := z.send.Uploader.UploadAttachment(filePath, threadID, threadType)
	if startedListener {
		_ = z.socket.StopListening()
		z.socket.SetUploadOnly(false)
	}
	return result, err
}

func (z *ZaloAPI) FetchAccountInfo() (any, error) {
	z.refreshServices()
	raw, err := z.get.FetchAccountInfo()
	if err == nil {
		z.cacheAccountProfile(raw)
	}
	return raw, err
}
func (z *ZaloAPI) FetchUserInfo(userIDs ...string) (any, error) {
	z.refreshServices()
	return z.get.FetchUserInfo(userIDs...)
}
func (z *ZaloAPI) FetchGroupInfo(groupIDs ...string) (any, error) {
	z.refreshServices()
	return z.get.FetchGroupInfo(groupIDs...)
}
func (z *ZaloAPI) FetchAllFriends() (any, error) { z.refreshServices(); return z.get.FetchAllFriends() }
func (z *ZaloAPI) FetchAllGroups() (any, error)  { z.refreshServices(); return z.get.FetchAllGroups() }
func (z *ZaloAPI) ListFriendRequests() (any, error) {
	z.refreshServices()
	return z.get.ListFriendRequests()
}
func (z *ZaloAPI) GetGroupMember(threadID string) (any, error) {
	z.refreshServices()
	return z.get.GetGroupMember(threadID)
}
func (z *ZaloAPI) GetRecentGroup(groupID string) (any, error) {
	z.refreshServices()
	return z.group.GetRecentGroup(groupID)
}
func (z *ZaloAPI) FetchPhoneNumber(phoneNumber string, language string) (*worker.User, error) {
	z.refreshServices()
	return z.get.FetchPhoneNumber(phoneNumber, language)
}
func (z *ZaloAPI) FindSticker(keyword string) (any, error) {
	z.refreshServices()
	return z.get.FindSticker(keyword)
}
func (z *ZaloAPI) GetAvatar(userID string, avatarSize int) (any, error) {
	z.refreshServices()
	return z.get.GetAvatar(userID, avatarSize)
}
func (z *ZaloAPI) GetCategory(catesID any) (any, error) {
	z.refreshServices()
	return z.get.GetCategory(catesID)
}
func (z *ZaloAPI) GetSticker(stickerID int) (any, error) {
	z.refreshServices()
	return z.get.GetSticker(stickerID)
}
func (z *ZaloAPI) SearchSticker(keyword string, limit int) (any, error) {
	z.refreshServices()
	return z.get.SearchSticker(keyword, limit)
}
func (z *ZaloAPI) UpdatePersonalSticker(cateIDs any, version int) (any, error) {
	z.refreshServices()
	return z.get.UpdatePersonalSticker(cateIDs, version)
}

func (z *ZaloAPI) AcceptFriendRequest(userID string, language string) (*worker.User, error) {
	z.refreshServices()
	return z.properties.AcceptFriendRequest(userID, language)
}
func (z *ZaloAPI) ChangeAccountAvatar(filePath string, width int, height int, language string, size int64) (*worker.User, error) {
	z.refreshServices()
	return z.properties.ChangeAccountAvatar(filePath, width, height, language, size)
}
func (z *ZaloAPI) ChangeAccountSetting(name string, dob string, gender int, biz map[string]any, language string) (*worker.User, error) {
	z.refreshServices()
	return z.properties.ChangeAccountSetting(name, dob, gender, biz, language)
}
func (z *ZaloAPI) MarkAsDelivered(msgID any, cliMsgID any, senderID any, threadID string, threadType core.ThreadType, method string) (bool, error) {
	z.refreshServices()
	return z.properties.MarkAsDelivered(msgID, cliMsgID, senderID, threadID, threadType, method)
}
func (z *ZaloAPI) MarkAsRead(msgID any, cliMsgID any, senderID any, threadID string, threadType core.ThreadType, method string) (bool, error) {
	z.refreshServices()
	return z.properties.MarkAsRead(msgID, cliMsgID, senderID, threadID, threadType, method)
}
func (z *ZaloAPI) SetTyping(threadID string, threadType core.ThreadType) (bool, error) {
	z.refreshServices()
	return z.properties.SetTyping(threadID, threadType)
}
func (z *ZaloAPI) UndoMessage(msgID any, cliMsgID any, threadID string, threadType core.ThreadType) (any, error) {
	z.refreshServices()
	return z.properties.UndoMessage(msgID, cliMsgID, threadID, threadType)
}

func (z *ZaloAPI) AddUsersToGroup(userIDs any, groupID string) (*worker.Group, error) {
	z.refreshServices()
	return z.group.AddUsersToGroup(userIDs, groupID)
}
func (z *ZaloAPI) ChangeGroupAvatar(filePath string, groupID string) (*worker.Group, error) {
	z.refreshServices()
	return z.group.ChangeGroupAvatar(filePath, groupID)
}
func (z *ZaloAPI) ChangeGroupName(groupName string, groupID string) (*worker.Group, error) {
	z.refreshServices()
	return z.group.ChangeGroupName(groupName, groupID)
}
func (z *ZaloAPI) ChangeGroupOwner(newAdminID string, groupID string) (*worker.Group, error) {
	z.refreshServices()
	return z.group.ChangeGroupOwner(newAdminID, groupID)
}
func (z *ZaloAPI) ChangeGroupSetting(groupID string, defaultMode string, kwargs map[string]any) (*worker.Group, error) {
	z.refreshServices()
	return z.group.ChangeGroupSetting(groupID, defaultMode, kwargs)
}
func (z *ZaloAPI) CheckGroup(link string) map[string]any {
	z.refreshServices()
	return z.group.CheckGroup(link)
}
func (z *ZaloAPI) CreateGroup(name string, description string, members any, nameChanged int, createLink int) (any, error) {
	z.refreshServices()
	return z.group.CreateGroup(name, description, members, nameChanged, createLink)
}
func (z *ZaloAPI) GetBlockedMembers(grid string, page int, count int) map[string]any {
	z.refreshServices()
	return z.group.GetBlockedMembers(grid, page, count)
}
func (z *ZaloAPI) GetGroupLink(threadID string) map[string]any {
	z.refreshServices()
	return z.group.GetGroupLink(threadID)
}
func (z *ZaloAPI) GetLastMsgs() (*worker.User, error) {
	z.refreshServices()
	return z.group.GetLastMsgs()
}
func (z *ZaloAPI) GetQRLink(userID string) (any, error) {
	z.refreshServices()
	return z.group.GetQRLink(userID)
}
func (z *ZaloAPI) UpgradeCommunity(groupID string) (*worker.Group, error) {
	z.refreshServices()
	return z.group.UpgradeCommunity(groupID)
}
func (z *ZaloAPI) GetGroupBoardList(groupID string, page int, count int, lastID int64, lastType int) (*worker.Group, error) {
	z.refreshServices()
	return z.group.GetGroupBoardList(groupID, page, count, lastID, lastType)
}
func (z *ZaloAPI) GetGroupPinMsg(groupID string, page int, count int, lastID int64, lastType int) (*worker.Group, error) {
	z.refreshServices()
	return z.group.GetGroupPinMsg(groupID, page, count, lastID, lastType)
}
func (z *ZaloAPI) GetGroupNote(groupID string, page int, count int, lastID int64, lastType int) (*worker.Group, error) {
	z.refreshServices()
	return z.group.GetGroupNote(groupID, page, count, lastID, lastType)
}
func (z *ZaloAPI) GetGroupPoll(groupID string, page int, count int, lastID int64, lastType int) (*worker.Group, error) {
	z.refreshServices()
	return z.group.GetGroupPoll(groupID, page, count, lastID, lastType)
}
func (z *ZaloAPI) AddAdmins(members any, groupID string) (*worker.Group, error) {
	z.refreshServices()
	return z.groupAction.AddAdmins(members, groupID)
}
func (z *ZaloAPI) BlockUsers(members any, groupID string) (*worker.Group, error) {
	z.refreshServices()
	return z.groupAction.BlockUsers(members, groupID)
}
func (z *ZaloAPI) CreatePoll(question string, options any, groupID string, expiredTime int64, pinAct bool, multiChoices bool, allowAddNewOption bool, hideVotePreview bool, isAnonymous bool) (*worker.Group, error) {
	z.refreshServices()
	return z.groupAction.CreatePoll(question, options, groupID, expiredTime, pinAct, multiChoices, allowAddNewOption, hideVotePreview, isAnonymous)
}
func (z *ZaloAPI) JoinGroup(inviteURL string) (any, error) {
	z.refreshServices()
	return z.groupAction.JoinGroup(inviteURL)
}
func (z *ZaloAPI) KickUsers(members any, groupID string) (*worker.Group, error) {
	z.refreshServices()
	return z.groupAction.KickUsers(members, groupID)
}
func (z *ZaloAPI) LeaveGroup(groupID string, silent bool) (any, error) {
	z.refreshServices()
	return z.groupAction.LeaveGroup(groupID, silent)
}
func (z *ZaloAPI) LockPoll(pollID int64) (*worker.Group, error) {
	z.refreshServices()
	return z.groupAction.LockPoll(pollID)
}
func (z *ZaloAPI) RemoveAdmins(members any, groupID string) (*worker.Group, error) {
	z.refreshServices()
	return z.groupAction.RemoveAdmins(members, groupID)
}
func (z *ZaloAPI) UnblockUsers(members any, groupID string) (*worker.Group, error) {
	z.refreshServices()
	return z.groupAction.UnblockUsers(members, groupID)
}
func (z *ZaloAPI) VotePoll(pollID int64, optionIDs any, groupID string) (any, error) {
	z.refreshServices()
	return z.groupAction.VotePoll(pollID, optionIDs, groupID)
}
func (z *ZaloAPI) DeleteMessage(msgID any, ownerID any, clientMsgID any, groupID string, onlyMe bool) (*worker.Group, error) {
	z.refreshServices()
	return z.groupMessage.DeleteMessage(msgID, ownerID, clientMsgID, groupID, onlyMe)
}
func (z *ZaloAPI) PinMessage(pinMsg *worker.MessageObject, groupID string) (*worker.Group, error) {
	z.refreshServices()
	return z.groupMessage.PinMessage(pinMsg, groupID)
}
func (z *ZaloAPI) UnpinMessage(pinID any, pinTime int64, groupID string) (*worker.Group, error) {
	z.refreshServices()
	return z.groupMessage.UnpinMessage(pinID, pinTime, groupID)
}
func (z *ZaloAPI) BoxInviteAccept(groupID string, lang string) (any, error) {
	z.refreshServices()
	return z.groupStatus.BoxInviteAccept(groupID, lang)
}
func (z *ZaloAPI) DisableLink(grid string) map[string]any {
	z.refreshServices()
	return z.groupStatus.DisableLink(grid)
}
func (z *ZaloAPI) DisperseGroup(groupID string) (*worker.Group, error) {
	z.refreshServices()
	return z.groupStatus.DisperseGroup(groupID)
}
func (z *ZaloAPI) GenerateNewLink(grid string) map[string]any {
	z.refreshServices()
	return z.groupStatus.GenerateNewLink(grid)
}
func (z *ZaloAPI) HandleGroupPending(members any, groupID string, isApprove bool) (*worker.Group, error) {
	z.refreshServices()
	return z.groupStatus.HandleGroupPending(members, groupID, isApprove)
}
func (z *ZaloAPI) ListInviteBox(page int, invPerPage int, mcount int, lastGroupID any) (any, error) {
	z.refreshServices()
	return z.groupStatus.ListInviteBox(page, invPerPage, mcount, lastGroupID)
}
func (z *ZaloAPI) SetMute(groupID string, mute bool) (any, error) {
	z.refreshServices()
	return z.groupStatus.SetMute(groupID, mute)
}
func (z *ZaloAPI) UpdateAutoDeleteChat(ttl int, threadID string, threadType core.ThreadType) (any, error) {
	z.refreshServices()
	return z.groupStatus.UpdateAutoDeleteChat(ttl, threadID, threadType)
}
func (z *ZaloAPI) ViewGroupPending(groupID string) (*worker.Group, error) {
	z.refreshServices()
	return z.groupStatus.ViewGroupPending(groupID)
}
func (z *ZaloAPI) ViewPollDetail(pollID int64) (*worker.Group, error) {
	z.refreshServices()
	return z.groupStatus.ViewPollDetail(pollID)
}

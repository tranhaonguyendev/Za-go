package group

import (
	"github.com/nguyendev/zago/internal/util"
	"github.com/nguyendev/zago/internal/worker"
)

func (g *GroupAPI) boardRequest(boardType int, groupID string, page int, count int, lastID int64, lastType int) (*worker.Group, error) {
	enc, err := g.Encode(map[string]any{"group_id": groupID, "board_type": boardType, "page": page, "count": count, "last_id": lastID, "last_type": lastType, "imei": g.IMEIValue()})
	if err != nil {
		return nil, err
	}
	data, err := g.GetJSON("https://groupboard-wpa.chat.zalo.me/api/board/list", g.Query(map[string]any{"params": enc, "zpw_ver": 645, "zpw_type": g.APILoginType}))
	if err != nil {
		return nil, err
	}
	decoded, err := g.ParseRaw(data)
	if err != nil {
		return nil, err
	}
	return worker.GroupFromDict(util.AsMap(decoded)["data"]), nil
}

func (g *GroupAPI) GetGroupBoardList(groupID string, page int, count int, lastID int64, lastType int) (*worker.Group, error) {
	return g.boardRequest(0, groupID, page, count, lastID, lastType)
}
func (g *GroupAPI) GetGroupPinMsg(groupID string, page int, count int, lastID int64, lastType int) (*worker.Group, error) {
	return g.boardRequest(2, groupID, page, count, lastID, lastType)
}
func (g *GroupAPI) GetGroupNote(groupID string, page int, count int, lastID int64, lastType int) (*worker.Group, error) {
	return g.boardRequest(1, groupID, page, count, lastID, lastType)
}
func (g *GroupAPI) GetGroupPoll(groupID string, page int, count int, lastID int64, lastType int) (*worker.Group, error) {
	return g.boardRequest(3, groupID, page, count, lastID, lastType)
}

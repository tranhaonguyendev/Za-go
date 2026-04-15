package handle

func (s *SendAPI) SetAlias(friendID string, alias string) (any, error) {
	enc, err := s.Encode(map[string]any{"friendId": friendID, "alias": alias, "imei": s.IMEIValue()})
	if err != nil {
		return nil, err
	}
	q := s.Query(map[string]any{"zpw_ver": 677, "zpw_type": s.APILoginType, "params": enc})
	data, err := s.GetJSON("https://tt-alias-wpa.chat.zalo.me/api/alias/update", q)
	if err != nil {
		return nil, err
	}
	return s.ParseRaw(data)
}

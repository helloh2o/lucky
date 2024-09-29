package immsg

type Msg interface {
	String()
}

type BaseMsg struct {
	FromId   string `json:"from_id"`
	FromName string `json:"from_name"`
	Avatar   string `json:"avatar"`
	Raw      []byte `json:"data"`
	Url      string `json:"url"`
	Type     int    `json:"type"`
	TimeFmt  string `json:"time_fmt"`
}

func (m BaseMsg) String() string {
	return string(m.Raw)
}

type ConnectMsg struct {
	BaseMsg
	UserId string   `json:"user_id"`
	Groups []string `json:"groups"`
}

type PeerMsg struct {
	BaseMsg
	ToUser string `json:"to_user"`
}

type PeerGroupMsg struct {
	BaseMsg
	ToGroup string `json:"to_group"`
	AtList  string `json:"at_list"`
}

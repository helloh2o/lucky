package immsg

type Msg interface {
	String()
}

type BaseMsg struct {
	Raw     []byte `json:"data"`
	Url     string `json:"url"`
	TimeFmt string `json:"time_fmt"`
	Type    int    `json:"type"`
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
	From string `json:"from"`
	To   string `json:"to"`
}

type PeerGroupMsg struct {
	BaseMsg
	From    string `json:"from"`
	GroupId string `json:"group"`
	AtList  string `json:"at_list"`
}

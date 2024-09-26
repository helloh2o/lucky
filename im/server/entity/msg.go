package entity

type Msg struct {
	Raw     []byte `json:"data"`
	Url     string `json:"url"`
	TimeFmt string `json:"time_fmt"`
	Type    string `json:"type"`
}

type PeerMsg struct {
	Msg
	From string `json:"from"`
	To   string `json:"to"`
}

type PeerGroupMsg struct {
	Msg
	From   string `json:"from"`
	Group  string `json:"group"`
	AtList string `json:"at_list"`
}

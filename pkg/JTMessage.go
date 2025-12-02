package pkg

type JTMessage struct {
	MsgID      uint16 `json:"MsgID"`
	Prop       uint16 `json:"Prop"`
	TerminalNo string `json:"TerminalNo"`
	SeqNo      uint16 `json:"SeqNo"`
	PkgCount   uint16 `json:"PkgSize"`
	PkgNo      uint16 `json:"PkgNo"`
	Body       []byte `json:"Body"`
}

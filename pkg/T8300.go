package pkg

import "net"

type T8300 struct {
	JTMessage
	MsgId
}

func (h *T8300) OnMsg(conn net.Conn) {
	//TODO implement me
	panic("implement me")
}

func init() {
	codec := &T8300{}
	codec.MsgID = P8300
	RegisterCodec(codec)
}
func (h *T8300) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T8300) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T8300) GetMsgId() MsgId {
	return h.MsgID
}

package pkg

import "net"

type T0003 struct {
	JTMessage
}

func (h *T0003) OnMsg(conn net.Conn) {
	//TODO implement me
	panic("implement me")
}

func init() {

	codec := &T0003{}
	codec.MsgID = P0003
	RegisterCodec(codec)
}
func (h *T0003) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T0003) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0003) GetMsgId() MsgId {
	return h.MsgID
}

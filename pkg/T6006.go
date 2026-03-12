package pkg

import "net"

type T6006 struct {
	JTMessage
}

func (h *T6006) OnMsg(conn net.Conn) {
	//TODO implement me
	panic("implement me")
}

func init() {

	codec := &T6006{}
	codec.MsgID = P6006
	RegisterCodec(codec)
}
func (h *T6006) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T6006) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T6006) GetMsgId() MsgId {
	return h.MsgID
}

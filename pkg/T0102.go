package pkg

import "net"

/*
*

	终端注册
*/
type T0102 struct {
	JTMessage
}

func (h *T0102) OnMsg(conn net.Conn) {
	//TODO implement me
	panic("implement me")
}

func init() {
	codec := &T0102{}
	codec.MsgID = P0102
	RegisterCodec(codec)
}
func (h *T0102) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T0102) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0102) GetMsgId() MsgId {
	return h.MsgID
}

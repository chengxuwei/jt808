package pkg

import "net"

/*
*

	终端注册
*/
type T0704 struct {
	JTMessage
}

func (h *T0704) OnMsg(conn net.Conn) {
	//TODO implement me
	panic("implement me")
}

func init() {

	codec := &T0704{}
	codec.MsgID = P0704
	RegisterCodec(codec)
}
func (h *T0704) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T0704) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0704) GetMsgId() MsgId {
	return h.MsgID
}

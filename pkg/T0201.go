package pkg

import (
	"net"
)

/*
*
 */
type T0201 struct {
	JTMessage
}

func (h *T0201) OnMsg(conn net.Conn) {
	//TODO implement me
	panic("implement me")
}

func init() {

	codec := &T0201{}
	codec.MsgID = P0201
	RegisterCodec(codec)
}
func (h *T0201) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T0201) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0201) GetMsgId() MsgId {
	return h.MsgID
}

package pkg

import "net"

/*
*
 */
type T8001 struct {
	JTMessage
}

func (h *T8001) OnMsg(conn net.Conn) {
	//TODO implement me
	panic("implement me")
}

func init() {
	codec := &T8001{}
	codec.MsgID = P8001
	RegisterCodec(codec)
}
func (h *T8001) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T8001) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T8001) GetMsgId() MsgId {
	return h.MsgID
}

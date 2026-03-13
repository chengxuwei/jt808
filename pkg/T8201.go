package pkg

import "net"

/*
*
 */
type T8201 struct {
	JTMessage
}

func (h *T8201) OnMsg(conn net.Conn) {
	//TODO implement me
	panic("implement me")
}

func init() {
	codec := &T8201{}
	codec.MsgID = P8201
	RegisterDecode(codec)
	RegisterEncode(codec)
}
func (h *T8201) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T8201) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T8201) GetMsgId() MsgId {
	return h.MsgID
}

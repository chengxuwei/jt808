package pkg

import (
	"log/slog"
	"net"
)

/*
*

	终端通用应答
*/
type T0001 struct {
	JTMessage
}

func (h *T0001) OnMsg(conn net.Conn) {
	slog.Info("不处理设备通用应答")
}

func init() {

	codec := &T0001{}
	codec.MsgID = P0001
	RegisterDecode(codec)
	RegisterEncode(codec)
}
func (h *T0001) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body
	h.JTMessage = *msg
	return nil
}

func (h *T0001) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0001) GetMsgId() MsgId {
	return P0001
}

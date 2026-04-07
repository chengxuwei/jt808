package pkg

import (
	"log/slog"
	"net"
)

/*
*
 */
type T8201 struct {
	JTMessage
}

func (h *T8201) OnMsg(conn net.Conn) {
	// T8201 是平台下发的位置查询，终端收到后上报位置；注释原因：平台侧收到回调时无需处理；注释人：Cursor
	slog.Debug("收到 T8201 平台位置查询（下行消息，无需处理）", slog.String("terminalNo", h.TerminalNo))
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

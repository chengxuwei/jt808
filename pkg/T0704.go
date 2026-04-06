package pkg

import (
	"encoding/hex"
	"log/slog"
	"net"
)

/*
*

	终端注册
*/
type T0704 struct {
	JTMessage
}

func (h *T0704) OnMsg(conn net.Conn) {
	//包回昨
	frame1 := Get8001Buf(h.JTMessage, 0)
	slog.Info("【T0704】回复8001", slog.Any("frame", hex.EncodeToString(frame1)))
	conn.Write(frame1)
}

func init() {

	codec := &T0704{}
	codec.MsgID = P0704
	RegisterDecode(codec)
	RegisterEncode(codec)
}
func (h *T0704) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T0704) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0704) GetMsgId() MsgId {
	return h.MsgID
}

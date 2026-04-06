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
type T0102 struct {
	JTMessage
}

func (h *T0102) OnMsg(conn net.Conn) {
	logJT808DecodedJSON("收到终端鉴权 0x0102（JSON）", h)
	frame1 := Get8001Buf(h.JTMessage, 0)
	slog.Info("T0102发送T8001帧", slog.Any("frame", hex.EncodeToString(frame1)))
	conn.Write(frame1)
}

func init() {
	codec := &T0102{}
	codec.MsgID = P0102

	RegisterDecode(codec)
	RegisterEncode(codec)
}
func (h *T0102) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body
	h.JTMessage = *msg

	return nil
}

func (h *T0102) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0102) GetMsgId() MsgId {
	return h.MsgID
}

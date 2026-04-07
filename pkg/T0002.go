package pkg

import (
	"encoding/hex"
	"log/slog"
	"net"
)

/*
*

	P0002终端-心跳
*/
type T0002 struct {
	JTMessage
}

func (h *T0002) OnMsg(conn net.Conn) {
	logJT808DecodedJSON("收到终端心跳（JSON）", h)
	frame1, seqNo := Get8001Buf(h.JTMessage, 0)
	slog.Info("T0002 回复 T8001",
		slog.String("terminalNo", h.TerminalNo),
		slog.String("recvMsgId", h.MsgID.String()),
		slog.Uint64("seqNo", uint64(seqNo)),
		slog.String("frameHex", hex.EncodeToString(frame1)),
	)
	if _, err := conn.Write(frame1); err != nil {
		slog.Error("T0002 回复 T8001 失败", slog.String("terminalNo", h.TerminalNo), slog.Any("err", err))
	}
}

func init() {
	//注册 TODO ，保证协程安全，可在连接时动态创建
	//TODO 增加context上下文处理
	codec := &T0002{}
	codec.MsgID = P0002
	RegisterDecode(codec)
	RegisterEncode(codec)
}
func (h *T0002) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body
	h.JTMessage = *msg
	return nil
}

func (h *T0002) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0002) GetMsgId() MsgId {
	return h.MsgID
}

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
	frame1, seqNo := Get8001Buf(h.JTMessage, 0)
	slog.Info("T0704 回复 T8001",
		slog.String("terminalNo", h.TerminalNo),
		slog.String("recvMsgId", h.MsgID.String()),
		slog.Uint64("seqNo", uint64(seqNo)),
		slog.String("frameHex", hex.EncodeToString(frame1)),
	)
	if _, err := conn.Write(frame1); err != nil {
		slog.Error("T0704 回复 T8001 失败", slog.String("terminalNo", h.TerminalNo), slog.Any("err", err))
	}
}

func init() {

	codec := &T0704{}
	codec.MsgID = P0704
	RegisterDecode(codec)
	RegisterEncode(codec)
}
func (h *T0704) Parse(msg *JTMessage) error {
	h.JTMessage = *msg
	return nil
}

func (h *T0704) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0704) GetMsgId() MsgId {
	return h.MsgID
}

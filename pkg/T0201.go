package pkg

import (
	"encoding/hex"
	"log/slog"
	"net"
)

/*
*
 */
type T0201 struct {
	JTMessage
}

func (h *T0201) OnMsg(conn net.Conn) {
	frame1, seqNo := Get8001Buf(h.JTMessage, 0)
	slog.Info("T0201 回复 T8001",
		slog.String("terminalNo", h.TerminalNo),
		slog.String("recvMsgId", h.MsgID.String()),
		slog.Uint64("seqNo", uint64(seqNo)),
		slog.String("frameHex", hex.EncodeToString(frame1)),
	)
	if _, err := conn.Write(frame1); err != nil {
		slog.Error("T0201 回复 T8001 失败", slog.String("terminalNo", h.TerminalNo), slog.Any("err", err))
	}
}

func init() {

	codec := &T0201{}
	codec.MsgID = P0201
	RegisterDecode(codec)
	RegisterEncode(codec)
}
func (h *T0201) Parse(msg *JTMessage) error {
	h.JTMessage = *msg
	return nil
}

func (h *T0201) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0201) GetMsgId() MsgId {
	return h.MsgID
}

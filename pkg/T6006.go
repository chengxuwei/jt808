package pkg

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
)

// T6006 私有扩展消息 0x6006；
// Body 布局：Flag(1字节) + Content(剩余字节, UTF-8)；注释原因：与 T8300 上行方向结构对称；注释人：Cursor
type T6006 struct {
	JTMessage
	// Flag 标志位，含义由厂商协议定义；注释原因：body[0] 固定为标志字节；注释人：Cursor
	Flag uint8 `json:"flag"`
	// Content 文本内容，UTF-8 编码；注释原因：body[1:] 为可变长文本；注释人：Cursor
	Content string `json:"context"`
}

func init() {
	codec := &T6006{}
	codec.MsgID = P6006
	RegisterDecode(codec)
	RegisterEncode(codec)
}

// Parse 按 Flag(1B)+Content(nB,UTF-8) 解析 T6006 消息体；
// 注释原因：body 至少 1 字节（Flag），Content 可为空；注释人：Cursor
func (h *T6006) Parse(msg *JTMessage) error {
	h.JTMessage = *msg
	if len(msg.Body) < 1 {
		return fmt.Errorf("T6006 body 长度不足，期望≥1字节，实际=%d", len(msg.Body))
	}
	h.Flag = msg.Body[0]
	if len(msg.Body) > 1 {
		h.Content = string(msg.Body[1:])
	}
	slog.Info("收到 T6006 私有扩展消息（0x6006）",
		slog.String("terminalNo", h.TerminalNo),
		slog.String("msgId", h.MsgID.String()),
		slog.Uint64("recvSeqNo", uint64(h.SeqNo)),
		slog.Uint64("flag", uint64(h.Flag)),
		slog.String("content", h.Content),
	)
	return nil
}

func (h *T6006) OnMsg(conn net.Conn) {
	logJT808DecodedJSON("收到 T6006 私有扩展消息（JSON）", h)
	frame1, seqNo := Get8001Buf(h.JTMessage, 0)
	slog.Info("T6006 回复 T8001",
		slog.String("terminalNo", h.TerminalNo),
		slog.String("recvMsgId", h.MsgID.String()),
		slog.Uint64("recvSeqNo", uint64(h.SeqNo)),
		slog.Uint64("seqNo", uint64(seqNo)),
		slog.String("frameHex", hex.EncodeToString(frame1)),
	)
	if _, err := conn.Write(frame1); err != nil {
		slog.Error("T6006 回复 T8001 失败", slog.String("terminalNo", h.TerminalNo), slog.Any("err", err))
	}
}

func (h *T6006) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T6006) GetMsgId() MsgId {
	return h.MsgID
}

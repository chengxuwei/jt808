package pkg

import (
	"log/slog"
	"net"
)

// T8300 平台文本信息下发 0x8300（JT/T 808）
// Flag 标志位：bit0=紧急 bit1=手动关闭 bit2=终端显示 bit3=居中显示 bit4=定时播报；注释人：Cursor
type T8300 struct {
	JTMessage
	Flag    uint8  `json:"Flag"`    // 标志，JSON字段名保持 type 与协议工具兼容；注释人：Cursor
	Content string `json:"Content"` // 文本内容，UTF-8 编码
}

func (h *T8300) OnMsg(conn net.Conn) {
	// T8300 是平台下发的文本信息，终端收到后显示；注释原因：平台侧收到回调时无需处理；注释人：Cursor
	slog.Debug("收到 T8300 文本信息下发（下行消息，无需处理）", slog.String("terminalNo", h.TerminalNo))
}

func init() {
	codec := &T8300{}
	codec.MsgID = P8300
	RegisterDecode(codec)
	RegisterEncode(codec)
}

func (h *T8300) Parse(msg *JTMessage) error {
	h.JTMessage = *msg
	return nil
}

// Encode 编码 T8300 文本信息下发帧；注释原因：Body = 标志(1字节) + 文本内容(UTF-8)；注释人：Cursor
func (h *T8300) Encode() []byte {
	slog.Info("[下行-6] T8300 编码前内容",
		slog.String("terminalNo", h.TerminalNo),
		slog.Uint64("seqNo", uint64(h.SeqNo)),
		slog.Uint64("flag", uint64(h.Flag)),
		slog.String("content", h.Content),
		slog.Int("contentLen", len(h.Content)),
	)
	content := []byte(h.Content)
	// Body: 标志(1字节) + 文本内容(n字节)
	body := make([]byte, 1+len(content))
	body[0] = h.Flag
	copy(body[1:], content)
	return PackFrame(JTMessage{
		MsgID:      P8300,
		TerminalNo: h.TerminalNo,
		SeqNo:      h.SeqNo,
		Body:       body,
	})
}

func (h *T8300) GetMsgId() MsgId {
	return h.MsgID
}

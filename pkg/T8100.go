package pkg

import (
	"encoding/binary"
	"log/slog"
	"net"
)

/*
*
 */
type T8100 struct {
	JTMessage
	ResponseSeqNo uint16 `json:"ResponseSeqNo"`
	Status        uint8  `json:"Status"`
	Token         string `json:"Token"`
}

func (h *T8100) OnMsg(conn net.Conn) {
	// T8100 是平台下发的注册应答，终端收到后由终端处理；注释原因：平台侧无需再回调；注释人：Cursor
	slog.Debug("收到 T8100 平台注册应答（下行消息，无需处理）", slog.String("terminalNo", h.TerminalNo))
}

func init() {
	codec := &T8100{}
	codec.MsgID = P8100
	RegisterDecode(codec)
	RegisterEncode(codec)
}

func (h *T8100) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T8100) Encode() []byte {
	token := []byte(h.Token)
	body := make([]byte, len(token)+2+1)
	binary.BigEndian.PutUint16(body[0:2], h.ResponseSeqNo)
	body[2] = h.Status
	copy(body[3:], token)
	msg := JTMessage{
		TerminalNo: h.TerminalNo,
		SeqNo:      h.SeqNo,
		Body:       body,
		MsgID:      h.MsgID,
	}
	frame := PackFrame(msg)
	return frame
}

func (h *T8100) GetMsgId() MsgId {
	return h.MsgID
}

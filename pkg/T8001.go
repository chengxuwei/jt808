package pkg

import (
	"encoding/binary"
	"log/slog"
	"net"
)

/*
*
 */
type T8001 struct {
	JTMessage
	ResponseSeqNo uint16 `json:"ResponseSeqNo"`
	ResponseMsgID uint16 `json:"MsgID"`
	Result        uint8  `json:"Status"`
}

func (h *T8001) OnMsg(conn net.Conn) {
	// 平台通用应答由平台下发，终端收到后不需要平台再处理；注释原因：T8001 是下行消息；注释人：Cursor
	slog.Debug("收到 T8001 平台通用应答（下行消息，无需处理）", slog.String("terminalNo", h.TerminalNo))
}

func init() {
	codec := &T8001{}
	codec.MsgID = P8001
	RegisterDecode(codec)
	RegisterEncode(codec)
}
func (h *T8001) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T8001) Encode() []byte {
	body := make([]byte, 5)
	//回复序号
	binary.BigEndian.PutUint16(body[0:], h.ResponseSeqNo)
	//回复消息ID
	binary.BigEndian.PutUint16(body[2:], h.ResponseMsgID)
	//结果状态
	body[4] = h.Result
	msg := JTMessage{
		TerminalNo: h.TerminalNo,
		SeqNo:      h.SeqNo,
		Body:       body,
		MsgID:      h.MsgID,
	}
	frame := PackFrame(msg)
	return frame
}

func (h *T8001) GetMsgId() MsgId {
	return h.MsgID
}

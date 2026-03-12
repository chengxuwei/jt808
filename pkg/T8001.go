package pkg

import (
	"encoding/binary"
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
	//TODO implement me
	panic("implement me")
}

func init() {
	codec := &T8001{}
	codec.MsgID = P8001
	RegisterCodec(codec)
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
		SeqNo:      111,
		Body:       body,
		MsgID:      h.MsgID,
	}
	frame := PackFrame(msg)
	return frame
}

func (h *T8001) GetMsgId() MsgId {
	return h.MsgID
}

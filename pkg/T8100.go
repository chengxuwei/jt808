package pkg

import (
	"encoding/binary"
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
	//TODO implement me
	panic("implement me")
}

func init() {
	codec := &T8100{}
	codec.MsgID = P8100
	RegisterCodec(codec)
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

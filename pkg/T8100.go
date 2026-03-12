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
	token         string `json:"token"`
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
	token := []byte(h.token)
	buf := make([]byte, len(token)+2+1)
	binary.BigEndian.PutUint16(buf[0:2], h.ResponseSeqNo)
	buf[2] = h.Status
	copy(buf[3:], token)

	return buf
}

func (h *T8100) GetMsgId() MsgId {
	return h.MsgID
}

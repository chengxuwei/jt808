package pkg

/*
*

	终端通用应答
*/
type T0001 struct {
	JTMessage
}

func init() {

	RegisterDecoder(P0002, &T0002{})
}
func (h *T0001) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T0001) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0002) Name() string {
	return h.MsgID.String()
}

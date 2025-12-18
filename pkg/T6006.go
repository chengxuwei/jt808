package pkg

type T6006 struct {
	JTMessage
}

func init() {

	RegisterDecoder(P6006, &T6006{})
}
func (h *T6006) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T6006) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T6006) Name() string {
	return h.MsgID.String()
}

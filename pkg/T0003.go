package pkg

type T0003 struct {
	JTMessage
}

func init() {

	RegisterDecoder(P0003, &T0003{})
}
func (h *T0003) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T0003) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0003) Name() string {
	return h.MsgID.String()
}

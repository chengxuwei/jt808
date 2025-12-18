package pkg

type T8300 struct {
	JTMessage
}

func init() {

	RegisterDecoder(P8300, &T8300{})
}
func (h *T8300) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T8300) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T8300) Name() string {
	return h.MsgID.String()
}

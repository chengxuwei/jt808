package pkg

/*
*
 */
type T0201 struct {
	JTMessage
}

func init() {

	RegisterDecoder(P0201, &T0201{})
}
func (h *T0201) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T0201) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0201) Name() string {
	return h.MsgID.String()
}

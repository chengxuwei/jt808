package pkg

/*
*
 */
type T8201 struct {
	JTMessage
}

func init() {

	RegisterDecoder(P8201, &T8201{})
}
func (h *T8201) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T8201) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T8201) Name() string {
	return h.MsgID.String()
}

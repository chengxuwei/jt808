package pkg

/*
*
 */
type T8001 struct {
	JTMessage
}

func init() {

	RegisterDecoder(P8001, &T8001{})
}
func (h *T8001) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T8001) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T8001) Name() string {
	return h.MsgID.String()
}

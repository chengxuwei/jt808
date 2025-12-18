package pkg

/*
*

	终端注册
*/
type T0704 struct {
	JTMessage
}

func init() {

	RegisterDecoder(P0704, &T0704{})
}
func (h *T0704) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T0704) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0704) Name() string {
	return h.MsgID.String()
}

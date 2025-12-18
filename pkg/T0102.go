package pkg

/*
*

	终端注册
*/
type T0102 struct {
	JTMessage
}

func init() {
	//注册 TODO ，保证协程安全，可在连接时动态创建
	//TODO 增加context上下文处理
	RegisterDecoder(P0100, &T0102{})
}
func (h *T0102) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T0102) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0102) Name() string {
	return h.MsgID.String()
}

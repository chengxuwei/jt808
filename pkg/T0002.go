package pkg

/*
*

	终端通用应答
*/
type T0002 struct {
	JTMessage
}

func init() {
	//注册 TODO ，保证协程安全，可在连接时动态创建
	//TODO 增加context上下文处理
	RegisterDecoder(P0001, &T0001{})
}
func (h *T0002) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}

func (h *T0002) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0001) Name() string {
	return h.MsgID.String()
}

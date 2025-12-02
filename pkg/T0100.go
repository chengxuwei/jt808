package pkg

type T0100 struct {
	JTMessage
}

func init() {
	//注册 TODO ，保证协程安全，可在连接时动态创建
	//TODO 增加context上下文处理
	RegisterDecoder(P0100, &T0100{})
}
func (h *T0100) Parse(jtMsg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	return nil
}
func (h *T0100) Protocol() uint16 { return P0200 }
func (h *T0100) Encode() []byte   { return []byte{0x02, 0x00} }
func (h *T0100) OnReadMsg(jtMsg *JTMessage) {

}

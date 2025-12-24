package pkg

import "fmt"

/*
*

	终端注册
*/
type T0100 struct {
	JTMessage
	ProvinceId uint16 `json:"province_id"` //省份ID
	CityId     uint16 `json:"city_id"`     //城市ID

}

func init() {
	//注册 TODO ，保证协程安全，可在连接时动态创建
	//TODO 增加context上下文处理
	RegisterDecoder(P0100, &T0100{})
}
func (h *T0100) Parse(msg *JTMessage) error {
	// 假设 body 已经在 jtMsg.Body

	//len(msg.Body)
	// 检查body长度是否为0
	if len(msg.Body) != 0 {
		return fmt.Errorf("body长度错误，期望为0，实际为%d", len(msg.Body))
	}
	return nil
}

func (h *T0100) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0100) Name() string {
	return h.MsgID.String()
}

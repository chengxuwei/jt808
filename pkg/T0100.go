package pkg

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
)

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
	codec := &T0100{}
	codec.MsgID = P0100
	RegisterCodec(codec)
}
func (h *T0100) Parse(msg *JTMessage) error {
	h.JTMessage = *msg
	// 假设 body 已经在 jtMsg.Body
	slog.Info("解析数据，然后Process处理", slog.Any("msg", msg))
	//len(msg.Body)
	// 检查body长度是否为0
	if len(msg.Body) != 0 {
		return fmt.Errorf("body长度错误，期望为0，实际为%d", len(msg.Body))
	}

	return nil
}

func (h *T0100) Encode() []byte { return []byte{0x02, 0x00} }

func (h *T0100) GetMsgId() MsgId {
	return P0100
}
func (h *T0100) OnMsg(conn net.Conn) {

	//业务回复
	t8100 := T8100{
		JTMessage: JTMessage{
			MsgID:      P8100,
			TerminalNo: h.TerminalNo,
		},
		ResponseSeqNo: h.SeqNo,
		Status:        0,
		Token:         "htzj-1264579279",
	}
	frame := t8100.Encode()
	slog.Info("转义编码发送帧", slog.Any("frame", hex.EncodeToString(frame)))
	conn.Write(frame)

}
